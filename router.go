package ipcrouter

import (
	"github.com/dist-ribut-us/ipc"
	"github.com/dist-ribut-us/log"
	"github.com/dist-ribut-us/message"
	"github.com/dist-ribut-us/rnet"
	"time"
)

// Handler for a Base message
type Handler func(*Base)

// Router handles the logic of routing ipc messages
type Router struct {
	ipc        *ipc.Proc
	callbacks  *handlers
	services   *handlers
	NetHandler Handler
}

// New creates a router on a port
func New(port rnet.Port) (*Router, error) {
	proc, err := ipc.New(port)
	if err != nil {
		return nil, err
	}
	i := &Router{
		ipc:       proc,
		callbacks: newhandlers(),
		services:  newhandlers(),
	}
	proc.Handler(i.handler)
	return i, nil
}

// Run will start the listen loop. Calling run multiple times will not start
// multiple listen loop.
func (i *Router) Run() { i.ipc.Run() }

// IsRunning indicates if the listen loop is running
func (i *Router) IsRunning() bool { return i.ipc.IsRunning() }

// Port returns the UDP port
func (i *Router) Port() rnet.Port { return i.ipc.Port() }

// String returns the address of the process
func (i *Router) String() string { return i.ipc.String() }

// IsOpen returns true if the connection is open. If the server is closed, it
// can neither send nor receive
func (i *Router) IsOpen() bool { return i.ipc.IsOpen() }

// Stop will stop the server
func (i *Router) Stop() error { return i.ipc.Stop() }

// Close will close the connection, freeing the port
func (i *Router) Close() error { return i.ipc.Close() }

// Send exposes a raw method for sending
func (i *Router) Send(port rnet.Port, msg *message.Header) {
	id := msg.Id
	if port == i.ipc.Port() {
		go i.baseHandler(&Base{
			port:   port,
			proc:   i,
			Header: msg,
		})
		return
	}

	msg.Id = 0
	i.ipc.Send(id, msg.Marshal(), port)
}

// Register a handler with a service ID on the router
func (i *Router) Register(service uint32, handler Handler) {
	if _, isTaken := i.services.get(service); !isTaken {
		i.services.set(service, handler)
	}
}

func (i *Router) handler(pkg *ipc.Package) {
	log.Debug("got_package", i.Port())

	b := i.toBase(pkg)
	if b == nil {
		log.Info(log.Lbl("not_a_message"), pkg.Addr)
	}

	i.baseHandler(b)
}

func (i *Router) baseHandler(b *Base) {
	if b.IsResponse() {
		if handler, ok := i.callbacks.get(b.Id); ok {
			handler(b)
			return
		}
	}

	if b.IsToNet() && i.NetHandler != nil {
		i.NetHandler(b)
		return
	}

	if handler, ok := i.services.get(b.Service); ok {
		handler(b)
		return
	}

	log.Info(log.Lbl("no_handler_or_callback"), b.IsResponse(), b.Id, b.Service, b.port)
}

func (i *Router) removeCallback(id uint32) {
	time.Sleep(time.Second)
	i.callbacks.delete(id)
}
