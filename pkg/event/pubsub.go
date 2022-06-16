package event

type PubSub interface {
	// Publish sends input message to specified channels.
	Publish(topic string, message []byte) error
	// Subscribe opens channel to listen specified channels.
	Subscribe(topics ...string) (Subscription, error)

	Unsubscribe(topics ...string) error
	// Close stops the pubsub hub.
	Close() error
}
