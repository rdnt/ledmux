package socket

import (
	"context"

	"ledctl3/pkg/event"
	"ledctl3/pkg/pubsub"
)

var (
	Connected    = "connected"
	Disconnected = "disconnected"
	Error        = "error"
)

type Client interface {
	Id() string
	Send(b []byte) error
	Receive() ([]byte, error)
}

type Encoder func(v interface{}) ([]byte, error)

type Decoder func(b []byte, v interface{}) error

type EventHandler func(clientId string, e event.Event) error

type Router struct {
	dispatcher *Dispatcher
	pubsub     *pubsub.MessageBroker
	encode     Encoder
	decode     Decoder
	handle     EventHandler
}

func New(opts ...Option) *Router {
	r := &Router{
		dispatcher: NewDispatcher(),
		pubsub:     pubsub.New(),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

type Option func(r *Router)

func WithEncodeFunc(enc Encoder) Option {
	return func(r *Router) {
		r.encode = enc
	}
}

func WithDecodeFunc(dec Decoder) Option {
	return func(r *Router) {
		r.decode = dec
	}
}

func WithEventHandlerFunc(h EventHandler) Option {
	return func(r *Router) {
		r.handle = h
	}
}

func (r *Router) On(event string, h HandlerFunc) {
	r.dispatcher.On(event, h)
}

func (r *Router) SetEncoder(enc Encoder) {
	r.encode = enc
}

func (r *Router) SetDecoder(dec Decoder) {
	r.decode = dec
}

func (r *Router) SetEventHandler(h EventHandler) {
	r.handle = h
}

func (r *Router) NewContext(ctx context.Context, clientId string) *Context {
	return &Context{
		router:        r,
		id:            clientId,
		ctx:           ctx,
		subscriptions: []*pubsub.Subscription{},
	}
}
