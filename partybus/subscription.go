package partybus

var nextID int64

type SubscriptionId int64

type Subscription struct {
	id         SubscriptionId
	bus        *Bus
	sender     chan<- Event
	receiver   <-chan Event
	eventTypes []EventType
}

func newSubscription(bus *Bus, eventKinds []EventType) *Subscription {
	nextID++
	sender, receiver := newQueue()
	return &Subscription{
		id:         SubscriptionId(nextID),
		bus:        bus,
		sender:     sender,
		receiver:   receiver,
		eventTypes: eventKinds,
	}
}

func (s *Subscription) Unsubscribe() error {
	return s.bus.Unsubscribe(s)
}

func (s *Subscription) Events() <-chan Event {
	return s.receiver
}
