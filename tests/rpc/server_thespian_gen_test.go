// code generated by thespian; DO NOT EDIT

package super

import (
	"context"
	"github.com/djmitche/thespian"
)

// serverBase is embedded in the private actor struct and contains
// common fields as well as default method implementations
type serverBase struct {
	rt *thespian.Runtime
	tx *ServerTx
	rx *ServerRx
}

// handleStart is called when the actor starts.  The default implementation
// does nothing, but users may implement this method to perform startup.
func (a *serverBase) handleStart() {}

// handleStop is called when the actor stops cleanly.  The default
// implementation does nothing, but users may implement this method to perform
// cleanup.
func (a *serverBase) handleStop() {}

// handleSuperEvent is called for supervisory events.  Actors which do not
// supervise need not implement this method.
func (a *serverBase) handleSuperEvent(ev thespian.SuperEvent) {}

// ServerBuilder is used to build new Server actors.
type ServerBuilder struct {
	server
	lower StringRPCMailbox
	upper StringRPCMailbox
}

func (bldr ServerBuilder) spawn(rt *thespian.Runtime) *ServerTx {
	reg := rt.Register()
	bldr.lower.ApplyDefaults()
	bldr.upper.ApplyDefaults()

	rx := &ServerRx{
		id:         reg.ID,
		rt:         rt,
		stopChan:   reg.StopChan,
		superChan:  reg.SuperChan,
		healthChan: reg.HealthChan,
		lower:      bldr.lower.Rx(),
		upper:      bldr.upper.Rx(),
	}

	tx := &ServerTx{
		ID:       reg.ID,
		stopChan: reg.StopChan,
		lower:    bldr.lower.Tx(),
		upper:    bldr.upper.Tx(),
	}

	// copy to a new server instance
	pvt := bldr.server
	pvt.rt = rt
	pvt.rx = rx
	pvt.tx = tx

	go pvt.loop()
	return tx
}

// ServerRx contains the Rx sides of the mailboxes, for access from the
// Server implementation.
type ServerRx struct {
	id uint64
	rt *thespian.Runtime

	stopChan   <-chan struct{}
	superChan  <-chan thespian.SuperEvent
	healthChan <-chan struct{}
	lower      StringRPCRx
	upper      StringRPCRx
}

// supervise starts supervision of the actor identified by otherID.
// It is a shortcut to thespian.Runtime.Supervize.
func (rx *ServerRx) supervise(otherID uint64) {
	rx.rt.Supervise(rx.id, otherID)
}

// unsupervise stops supervision of the actor identified by otherID.
// It is a shortcut to thespian.Runtime.Unupervize.
func (rx *ServerRx) unsupervise(otherID uint64) {
	rx.rt.Unsupervise(rx.id, otherID)
}

// ServerTx is the public handle for Server actors.
type ServerTx struct {
	// ID is the unique ID of this actor
	ID       uint64
	stopChan chan<- struct{}
	lower    StringRPCTx
	upper    StringRPCTx
}

// Stop sends a message to stop the actor.  This does not wait until
// the actor has stopped.
func (a *ServerTx) Stop() {
	select {
	case a.stopChan <- struct{}{}:
	default:
	}
}

// Lower makes a synchronous request to the actor's lower mailbox.
func (tx *ServerTx) Lower(ctx context.Context, req string) (res string, err error) {
	resC := make(chan string)
	tx.lower.c <- StringRPCReq{ctx, req, resC}
	select {
	case res = <-resC:
		return
	case <-ctx.Done():
		err = ctx.Err()
		return
	}
}

// Upper makes a synchronous request to the actor's upper mailbox.
func (tx *ServerTx) Upper(ctx context.Context, req string) (res string, err error) {
	resC := make(chan string)
	tx.upper.c <- StringRPCReq{ctx, req, resC}
	select {
	case res = <-resC:
		return
	case <-ctx.Done():
		err = ctx.Err()
		return
	}
}

func (a *server) loop() {
	rx := a.rx
	defer func() {

		a.rt.ActorStopped(a.rx.id)
	}()
	a.handleStart()
	for {
		select {
		case <-rx.healthChan:
			// nothing to do
		case ev := <-rx.superChan:
			a.handleSuperEvent(ev)
		case <-rx.stopChan:
			a.handleStop()
			return
		case m := <-rx.lower.Chan():
			res := a.handleLower(m.ctx, m.req)
			m.resC <- res
		case m := <-rx.upper.Chan():
			res := a.handleUpper(m.ctx, m.req)
			m.resC <- res
		}
	}
}
