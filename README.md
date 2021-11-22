# Thespian

Thespian is a library supporting use of the [actor model](https://en.wikipedia.org/wiki/Actor_model) in Go code.

NOTE: _This is a work in progress and should not (yet? ever?) be used in real code._

## Introduction

Briefly, the actor model divides an application into independently-executing entities (actors) that respond to messages.
In response to a message, an actor may
 * send messages to other actors
 * create new actors
 * modify its private state to change the way it will react to subsequent messages

Notably, actors may *not* share data with other actors.
All communication occurs by message-passing.

## Actors and Mailboxes

A Thespian actor has
 * an instance of a private struct containing the actor's data:
 * mailboxes where incoming messages are received; and
 * `handleXxx` methods on the private type to respond to messages from each mailbox.

The library runs each actor in a dedicated Goroutine, and handles startup, shutdown, health monitoring, and other administrative details.

Mailboxes are a generalization of Go channels, and can provide:
 * simple message transfer between agents; and
 * time-related messages, such as a message every 15 seconds.
Future expansions might include
 * Complex message transfer with improved performance characteristics (such as memory re-use or batching multiple messages into one);
 * RPC message-passing, where the sender of the request message blocks waiting for a response message; and
 * Network listeners, where a new connection or data on an existing socket results in a message.

## Code Generation and Usage

This library implements actors by generating Go code based on a specification (`thespian.yml`).
Code generation provides typesafe, ergonomic access and avoids the performance overhead of vtable dispatch for interfaces.

The specification file results in code generated in the same directory.
It is typically invoked from a `go.gen` in the same directory.
For example:

```go
# go.gen
//go:generate go run github.com/djmitche/thespian/cmd/thespian generate
```

```yaml
# thespian.yml
actors:
  OrderTracker:
    mailboxes:
      newOrder:
        kind: simple
        message-type: "Order"
        type: Order
      orderComplete:
        kind: simple
        message-type: "Order"
        type: Order

mailboxes:
  Order:
    kind: simple
    message-type: PurchaseOrder
```

### Runtime

Actors run in the context of a Runtime, which tracks running actors and handles health-monitoring, supervision, and other oversight responsibilities.

Create a new Runtime with `thespian.NewRuntime()`.

### Actors

The `actors` property of the specification file describes the actor types that will be generated.
Each has a set of named mailboxes for that actor type.
Each mailbox specifies a kind and some kind-specific values.
These are described in the next section.

In addition to the specification in `thespian.yml`, you must supply a "private type" for the actor.
This type must begin by embedding the base type, and can contain any additional private data for the actor.
The type _must_ be private and access to an instance is limited to the agent it represents.
As such, no synchronization primitives (such as `sync.Mutex`) are required.

The private type must implement a `handleMailboxName` method for each mailbox.
Continuing the example above:

```go
type orderTracker struct {
    orderTrackerBase

    openOrders map[OrderID]Order
    closedOrders map[OrderID]Order
}

func (ot *orderTracker) handleNewOrder(msg Order) {
    // ...
}

func (ot *orderTracker) handleOrderComplete(msg Order) {
    // ...
}
```

The generated code contains several struct types, prefixed with the base name given in the specification.
For the "OrderTracker" type in the example, these are

 * `orderTrackerBase` - a base type that should be embedded in the private type, as above.
 * `OrderTrackerBuilder` - a builder for new actor instances
 * `OrderTrackerRx` - a struct to handle receiving messages from mailboxes (private to the actor)
 * `OrderTrackerTx` - a struct to handle sending messages to mailboxes (available to other actors)

#### Base Type

The `...Base` type provides default method implementations:

* `handleStart` - called on actor start
* `handleStop` - called on clean stop of an actor
* `handleSuperEvernt` - called for supervisory events

as well as fields:

* `rx` - pointer to the Rx instance for this actor, used to adjust mailbox behavior
* `tx` - pointer to the Tx instance for this actor, used to send messages to itself
* `rt` - pointer to the `thespian.Runtime` in which this actor is executing

#### Builder Type

The Builder type is used to build a new actor.
It contains an embedded private struct and a private field for each mailbox.
The embedded private struct can be used to set initial values for the actor, and the mailbox fields can be used to configure mailboxes before startup.
For example, a mailbox can be configured to be disabled at startup.

You should wrap the builder with one or more constructor functions, returning the `...Tx` type, such as

```go
func NewOrderTracker(rt *thespian.Runtime) *OrderTrackerTx {
    return OrderTrackerBuilder{
        orderTracker: {
            openOrders: make(map[OrderID]Order),
            closedOrders: make(map[OrderID]Order),
        },
        orderComplete: OrderMailbox {
            Disabled: true, // orderComplete mailbox will begin in a disabled state
        }
    }.spawn(rt)
}
```

#### Rx Type

The Rx type is the actor's interface to its mailboxes.
Most mailboxes allow some kind of runtime configuration.
For example, simple mailboxes can be enabled or disabled.
The Rx type has a field for each mailbox, of the mailbox's Rx type.

For example:
```go
func (ot *orderTracker) handleNewOrder(msg Order) {
    // now that we have an order, allow order completion messages
    ot.rx.orderComplete.Disabled = false
}
```

#### Tx Type

The Tx type is the public interface for an actor.
It contains only one public field (ID), and implements a method for each mailbox to which messages can be sent.

For example:
```go
ot := NewOrderTracker(rt)
ot.NewOrder(Order{ .. })
```

An instance of the Tx type may be passed around to any actor that wishes to send messages the actor.

The Tx type also contains a `Stop()` message which requests that the actor stop on its next iteration.

### Mailboxes

Mailboxes are generated from elements in the `mailboxes` property of `thespian.yml`.
The library also provides a few pre-generated mailbox implementations for common types and purposes.

The library defines several "kinds" of mailboxes, each described below.
Most define three types, each with suffixes of the base type name.
The Mailbox type defines the mailbox, and when an actor is spawned that Mailbox is split into an Rx and Tx instance.
From the example above, these would be

* `OrderMailbox`
* `OrderRx`
* `OrderTx`

#### Simple Mailboxes

Simple mailboxes simply wrap a typed Go channel.
They are defined like this in the specification file:

```yaml
mailboxes:
  Order:
    kind: simple
    message-type: PurchaseOrder
```

Where `message-type` is the type of the messages carried by the channel.
At the moment, this type must be defined in the same Go package.

Simple mailboxes can be used in actors as follows:

```yaml
actors:
  SomeActor:
    mailboxes:
      mailboxName:
        kind: simple
        message-type: "Order"
        import: my.package/path/to/mailboxes
        type: Order
```

Here, `message-type` must match the message type used in the mailbox specification, and `type` must match the base name of the mailbox type in the package identified by `import`.
The `import` property may be omitted if it is the same as the package in which the actor is defined.

The Mailbox type of a simple mailbox has a `C` field giving the channel that will carry the messages.
When building an actor, setting this field to a channel used by another actor instances will cause the actors to both read from the same channel, with the result that any message sent will reach only one of the waiting actors.
The default channel size is 10, but this can be overridden when building an actor by creating a channel of the desired size.

The Mailbox type also has a `Disabled` property.
Setting this to true will cause the Mailbox to start in the disabled state, meaning that it will not receive messages.
The default is to begin receiving messages at startup.

The Rx type has a Disabled property as well, that can be manipulated from the actor implementation.
In the example above, the OrderCompleted mailbox is enabled only after an order has been created.

#### Ticker Mailboxes

The `ticker` kind generates a mailbox that embeds a `time.Ticker`.
There are no Mailbox or Tx types for this mailbox.

When the actor starts, the ticker is disabled.
The ticker can be started with `rx.<mailboxName>.Reset(dur)`, and stopped with `rx.<mailboxName>.Stop()`.
On each tick, the `handleMailboxName(t time.Time)` method will be called.

## Supervision

(TBD)

An actor may call `rx.supervise(someID)` to begin "supervising" another actor.
The `rx.unsupervise(someID)` method does the reverse.

A supervising actor receives calls to `handleSuperEvent` when the state of the supervised actor changes.
The supported event types are:

 * `thespian.UnhealthyActor` - produced when a healthy actor becomes unhealthy
 * `thespian.HealthyActor` - produced when an unhealthy actor becomes healthy
 * `thespian.StoppedActor` - produced when the actor stops (whether cleanly or by panic)

The runtime monitors each actor to ensure that it is waiting for messages at least once per second.
When this check fails (such as when the actor is deadlocked, or spends too much time in a `handle` method), any supervising actors are notified.

## Caveats

Thespian does not guarantee the order in which messages are delivered between mailboxes.
In the "Orders" example above, it is possible for an OrderCompleted message to be delivered before the NewOrder message that created the corresponding order.
