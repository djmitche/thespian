package thespian

import (
	"fmt"
	"sync"
	"time"
)

type SuperEventType string

const (

	// UnhealthyActor is sent when the actor with the given ID becomes unhealthy
	UnhealthyActor SuperEventType = "unhealthy"

	// UnhealthyActor is sent when the actor with the given ID becomes healthy
	HealthyActor SuperEventType = "healthy"

	// StoppedActor is sent when the actor with the given ID stops
	StoppedActor SuperEventType = "stopped"
)

// A Thespian runtime manages a collection of actors.
type Runtime struct {
	// Mutex covers all fields of this type
	sync.Mutex

	// nextID is the next ID that will be handed out
	nextID uint64

	// actors is the collection of runtime's references to "live" actors
	actors map[uint64]*runtimeActor

	// stopped is signalled when the last actor stops
	stopped *sync.Cond
}

// Registration is returned from Register
type Registration struct {
	// ID is the identifier of the newly registered actor
	ID uint64

	// StopChan is a channel to stop the actor
	StopChan chan struct{}

	// SuperChan is a channel for supervisory messages
	SuperChan chan SuperEvent

	// HealthChan is the receive side of a channel for monitoring
	// actor health
	HealthChan <-chan struct{}
}

// SuperEvent represents a supervisory event.
type SuperEvent struct {
	Event SuperEventType
	ID    uint64
}

type runtimeActor struct {
	id         uint64
	stopChan   chan<- struct{}
	healthChan chan<- struct{}
	healthy    bool
	superChan  chan<- SuperEvent
	superSubs  map[uint64]struct{}
}

// NewRuntime creates a new runtime.
func NewRuntime() *Runtime {
	rt := &Runtime{
		nextID: 1,
		actors: make(map[uint64]*runtimeActor),
	}
	rt.stopped = sync.NewCond(&rt.Mutex)
	return rt
}

func (rt *Runtime) Register() Registration {
	stopChan := make(chan struct{}, 1)
	superChan := make(chan SuperEvent, 10)
	healthChan := make(chan struct{}, 2)

	rta := &runtimeActor{
		stopChan:   stopChan,
		healthChan: healthChan,
		healthy:    true,
		superChan:  superChan,
		superSubs:  map[uint64]struct{}{},
	}

	var id uint64
	func() {
		rt.Lock()
		defer rt.Unlock()

		id = rt.nextID
		rt.nextID++

		rt.actors[id] = rta
		rta.id = id
	}()

	return Registration{
		ID:         id,
		StopChan:   stopChan,
		SuperChan:  superChan,
		HealthChan: healthChan,
	}
}

// Inform the runtime that this actor has stopped.  This is called
// automatically when an actor finishes.
func (rt *Runtime) ActorStopped(id uint64) {
	rt.Lock()
	defer rt.Unlock()

	actor, found := rt.actors[id]
	if !found {
		panic(fmt.Sprintf("Actor %d stopped more than once", id))
	}

	rt.sendSuperEventLocked(actor, StoppedActor)

	delete(rt.actors, id)

	// if that was the last actor, signal any waiter
	if len(rt.actors) == 0 {
		rt.stopped.Broadcast()
	}
}

// Start the runtime and return immediately.
func (rt *Runtime) Start() {
	go rt.loop()
}

// Run starts the runtime and blocks until it stops
func (rt *Runtime) Run() {
	rt.Start()
	rt.wait()
}

// Stop the runtime gracefully, by stopping all actors and optionally waiting
// until they complete.
func (rt *Runtime) Stop(wait bool) {
	func() {
		rt.Lock()
		defer rt.Unlock()

		for _, actor := range rt.actors {
			select {
			case actor.stopChan <- struct{}{}:
			default:
			}
		}
	}()

	if wait {
		rt.wait()
	}
}

// Supervise subscribes the first actor to supervisory events about the
// second.  The simpler shortcut to subscribe the current actor to another is
// `actor.rx.supervise(otherID)`.
func (rt *Runtime) Supervise(superID, otherID uint64) {
	rt.Lock()
	defer rt.Unlock()

	_, found := rt.actors[superID]
	if !found {
		return
	}
	other, found := rt.actors[otherID]
	if !found {
		return
	}
	other.superSubs[superID] = struct{}{}
}

// Unsupervise unnsubscribes the first actor from supervisory events about the
// second.  The simpler shortcut to unsubscribe the current actor to another
// is `actor.rx.unsupervise(otherID)`.
func (rt *Runtime) Unsupervise(superID, otherID uint64) {
	rt.Lock()
	defer rt.Unlock()

	other, found := rt.actors[otherID]
	if !found {
		return
	}
	delete(other.superSubs, superID)
}

// wait blocks until there are no running actors
func (rt *Runtime) wait() {
	rt.Lock()
	defer rt.Unlock()

	for len(rt.actors) > 0 {
		rt.stopped.Wait()
	}
}

func (rt *Runtime) loop() {
	checker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-checker.C:
			rt.pingActors()
		}
	}
}

func (rt *Runtime) pingActors() {
	rt.Lock()
	defer rt.Unlock()

	for _, actor := range rt.actors {
		select {
		case actor.healthChan <- struct{}{}:
			// the actor is healthy
			if !actor.healthy {
				rt.sendSuperEventLocked(actor, HealthyActor)
				actor.healthy = true
			}
		default:
			// health chan is not being read from, so the actor is
			// likely unhealthy
			if actor.healthy {
				rt.sendSuperEventLocked(actor, UnhealthyActor)
				actor.healthy = false
			}
		}
	}
}

// Send a SuperEvent about the givne agent, with the runtime already locked.
func (rt *Runtime) sendSuperEventLocked(actor *runtimeActor, evt SuperEventType) {
	for subId := range actor.superSubs {
		sub, found := rt.actors[subId]
		if found {
			// TODO: nonblocking send with warning?
			sub.superChan <- SuperEvent{
				Event: evt,
				ID:    actor.id,
			}
		} else {
			delete(rt.actors, subId)
		}
	}
}
