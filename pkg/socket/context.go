package socket

import (
	"context"
	"fmt"
	"time"

	"ledctl3/pkg/event"
	"ledctl3/pkg/pubsub"
)

type Context struct {
	ctx           context.Context
	id            string
	keys          map[string]string
	data          []byte
	router        *Router
	subscriptions []*pubsub.Subscription
}

type HandlerFunc func(*Context)

func (c *Context) String() string {
	return fmt.Sprintf("{ keys: %#v, data: %#v }", c.keys, c.data)
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	if c.ctx == nil {
		return
	}

	return c.ctx.Deadline()
}

func (c *Context) Done() <-chan struct{} {
	if c.ctx == nil {
		return nil
	}

	return c.ctx.Done()
}

func (c *Context) Err() error {
	if c.ctx == nil {
		return nil
	}

	return c.ctx.Err()
}

func (c *Context) Value(key interface{}) interface{} {
	if keyAsString, ok := key.(string); ok {
		if val, exists := c.Get(keyAsString); exists {
			return val
		}
	}

	if c.ctx == nil {
		return nil
	}

	return c.ctx.Value(key)
}

func (c *Context) Bind(v interface{}) error {
	return c.router.decode(c.data, v)
}

func (c *Context) Data() []byte {
	return c.data
}

func (c *Context) Get(key string) (value string, exists bool) {
	if c.keys == nil {
		return "", false
	}

	value, exists = c.keys[key]

	return
}

func (c *Context) Set(key string, value string) {
	if c.keys == nil {
		c.keys = map[string]string{}
	}

	c.keys[key] = value
}

func (c *Context) Id() string {
	return c.id
}

func (c *Context) Join(channel string) {
	sub := c.router.pubsub.Subscribe(
		channel, c.Id(),
	)

	c.subscriptions = append(c.subscriptions, sub)

	go func() {
		for b := range sub.Events() {
			//log.Printf("received topic %s, sub %s, payload: %s", channel, sub.Id(), string(b))
			if c.Err() != nil {
				break
			}

			err := c.router.handle(c.Id(), b)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}()
}

func (c *Context) send(e string, v interface{}) {
	b, err := c.router.encode(v)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = c.router.handle(
		c.Id(), event.Event{
			Event: e,
			Data:  b,
		},
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

func (c *Context) Broadcast(topic string, e string, v interface{}) {
	b, err := c.router.encode(v)
	if err != nil {
		fmt.Println(err)
		return
	}

	// dispatch event only on contexes of other clients
	c.router.pubsub.Publish(
		topic, c.Id(), event.Event{
			Event: e,
			Data:  b,
		},
	)
}

func (c *Context) Success(e string, v interface{}) {
	c.send(e, v)
}

func (c *Context) Error(err error, v ...interface{}) {
	if len(v) == 0 {
		c.send(err.Error(), nil)
		return
	}

	c.send(err.Error(), v)
}

func (c *Context) Dispose() {
	for _, sub := range c.subscriptions {
		sub.Dispose()
	}
}

func (c *Context) Handle(e event.Event) {
	c.router.dispatcher.Dispatch(c, e.Event, e.Data)
}
