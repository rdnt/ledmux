package event

type Subscription interface {
	// Messages returns channel to receive messages.
	Messages() <-chan []byte
	// Close stops listening underlying pubsub topics.
	Close() error
	// Done returns channel to receive event when this channel is closed.
	Done() <-chan bool
}
