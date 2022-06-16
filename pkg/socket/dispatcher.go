package socket

import "fmt"

type Dispatcher struct {
	events map[string][]HandlerFunc
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		events: make(map[string][]HandlerFunc, 0),
	}
}

func (d *Dispatcher) On(name string, h HandlerFunc) {
	_, ok := d.events[name]
	if !ok {
		d.events[name] = make([]HandlerFunc, 0, 1)
	}

	d.events[name] = append(d.events[name], h)

	//log.Printf("added listener for event %s", name)
}

func (d *Dispatcher) Dispatch(c *Context, name string, b []byte) {
	handlers, ok := d.events[name]
	if !ok {
		fmt.Printf("%s event is not registered\n", name)
		return
	}

	c.data = b

	//log.Printf("triggering event %s, data %s", name, string(b))

	for _, h := range handlers {
		h(c)
	}
}
