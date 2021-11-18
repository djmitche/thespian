package thespian

import (
	"fmt"
	"log"
	"sync"
	"time"
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

func NewRuntime() *Runtime {
	rt := &Runtime{
		nextID: 1,
		actors: make(map[uint64]*runtimeActor),
	}
	rt.stopped = sync.NewCond(&rt.Mutex)
	return rt
}

func (rt *Runtime) Register(base *ActorBase) {
	if base.ID != 0 {
		panic("ActorBase has already been registered")
	}

	stopChan := make(chan struct{}, 1)
	healthChan := make(chan struct{}, 2)

	rta := &runtimeActor{stopChan, healthChan}

	func() {
		rt.Lock()
		defer rt.Unlock()

		base.ID = rt.nextID
		rt.nextID++

		rt.actors[base.ID] = rta
	}()

	base.Runtime = rt
	base.StopChan = stopChan
	base.HealthChan = healthChan
}

// Inform the runtime that this actor has stopped.
func (rt *Runtime) ActorStopped(base *ActorBase) {
	rt.Lock()
	defer rt.Unlock()

	_, found := rt.actors[base.ID]
	if !found {
		panic(fmt.Sprintf("Actor %d stopped more than once", base.ID))
	}
	log.Printf("Actor %d stopped", base.ID)
	delete(rt.actors, base.ID)

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
	log.Printf("pinging")
	unhealthy := []uint64{}
	func() {
		rt.Lock()
		defer rt.Unlock()

		for id, actor := range rt.actors {
			select {
			case actor.healthChan <- struct{}{}:
				continue
			default:
				// health chan is not being read from, so the actor is
				// likely unhealthy
				unhealthy = append(unhealthy, id)
			}
		}
	}()

	for id := range unhealthy {
		log.Printf("Actor %d is unhealthy\n", id) // TODO: react somehow
	}
}

type runtimeActor struct {
	stopChan   chan<- struct{}
	healthChan chan<- struct{}
}
