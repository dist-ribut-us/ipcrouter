package ipcrouter

import (
	"github.com/dist-ribut-us/errors"
	"github.com/dist-ribut-us/ipc"
	"github.com/dist-ribut-us/log"
	"github.com/dist-ribut-us/message"
	"github.com/dist-ribut-us/rnet"
	"time"
)

// Router handles the logic of routing ipc messages
type Router struct {
	ipc                *ipc.Proc
	callbacks          *callbacks
	netcallbacks       *netcallbacks
	queryServices      *queryServices
	commandServices    *commandServices
	netQueryServices   *netQueryServices
	netCommandServices *netCommandServices
	netSenderService   NetSenderService
	NetSenderPort      rnet.Port
}

// New creates a router on a port
func New(port rnet.Port) (*Router, error) {
	proc, err := ipc.New(port)
	if err != nil {
		return nil, err
	}
	i := &Router{
		ipc:                proc,
		callbacks:          newcallbacks(),
		netcallbacks:       newnetcallbacks(),
		queryServices:      newqueryServices(),
		commandServices:    newcommandServices(),
		netQueryServices:   newnetQueryServices(),
		netCommandServices: newnetCommandServices(),
	}
	proc.Handler(i.handler)
	return i, nil
}

// Run will start the listen loop. Calling run multiple times will not start
// multiple listen loop.
func (i *Router) Run() { i.ipc.Run() }

// IsRunning indicates if the listen loop is running
func (i *Router) IsRunning() bool { return i.ipc.IsRunning() }

// GetPort returns the UDP port
func (i *Router) GetPort() rnet.Port { return i.ipc.GetPort() }

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
	if port == i.ipc.GetPort() {
		go i.baseHandler(&base{
			port:   port,
			proc:   i,
			Header: msg,
		})
		return
	}
	msg.Id = 0
	i.ipc.Send(id, msg.Marshal(), port)
}

// Errors that can be returned when registering a service
const (
	ErrQueryServiceTaken      = errors.String("QueryService already registered")
	ErrCommandServiceTaken    = errors.String("CommandService already registered")
	ErrNetQueryServiceTaken   = errors.String("NetQueryService already registered")
	ErrNetCommandServiceTaken = errors.String("NetCommandService already registered")
	ErrNothingRegisered       = errors.String("Did not register any services")
)

// Register a handler with a service ID on the router
func (i *Router) Register(service Service) error {
	sid := service.ServiceID()
	log.Info(sid)
	registered := false

	if nss, ok := service.(NetSenderService); ok {
		i.netSenderService = nss
		if p, ok := nss.(rnet.Porter); ok {
			i.NetSenderPort = p.GetPort()
		}
		registered = true
	}

	if qs, ok := service.(QueryService); ok {
		if _, taken := i.queryServices.get(sid); taken {
			return ErrQueryServiceTaken
		}
		i.queryServices.set(sid, qs)
		registered = true
	}

	if cs, ok := service.(CommandService); ok {
		if _, taken := i.commandServices.get(sid); taken {
			return ErrCommandServiceTaken
		}
		i.commandServices.set(sid, cs)
		registered = true
	}

	if nqs, ok := service.(NetQueryService); ok {
		log.Info("NetQueryService", service.ServiceID(), i.GetPort())
		if _, taken := i.netQueryServices.get(sid); taken {
			return ErrNetQueryServiceTaken
		}
		i.netQueryServices.set(sid, nqs)
		registered = true
	}

	if ncs, ok := service.(NetCommandService); ok {
		log.Info("NetCommandService")
		if _, taken := i.netCommandServices.get(sid); taken {
			return ErrNetCommandServiceTaken
		}
		i.netCommandServices.set(sid, ncs)
		registered = true
	}

	if !registered {
		return ErrNothingRegisered
	}
	return nil
}

func (i *Router) handler(pkg *ipc.Package) {
	log.Debug("got_package", i.GetPort())

	b := i.toBase(pkg)
	if b == nil {
		log.Info(log.Lbl("not_a_message"), pkg.Addr)
	}

	i.baseHandler(b)
}

func (i *Router) baseHandler(b *base) {
	if b.IsResponse() {
		if b.IsFromNet() {
			if handler, ok := i.netcallbacks.get(b.Id); ok {
				handler(netResponse{b})
				return
			}
		} else if handler, ok := i.callbacks.get(b.Id); ok {
			handler(response{b})
			return
		}
	}

	if b.IsToNet() {
		b.UnsetFlag(message.ToNet)
		if i.netSenderService != nil {
			i.netSenderService.NetSend(netSendRequest{b})
		} else {
			log.Info(log.Lbl("send_no_net_service"), b.IsResponse(), b.Id, b.Service, b.port)
		}
		return
	}

	if b.IsFromNet() {
		if b.IsQuery() {
			if handler, ok := i.netQueryServices.get(b.Service); ok {
				handler.NetQueryHandler(netQuery{b})
				return
			}
		} else {
			if handler, ok := i.netCommandServices.get(b.Service); ok {
				handler.NetCommandHandler(netCommand{b})
				return
			}
		}
		log.Info(log.Lbl("no_net_handler_or_callback"), b.IsResponse(), b.IsQuery(), b.Id, b.Service, b.port, i.GetPort())
	}

	if b.IsQuery() {
		if qs, ok := i.queryServices.get(b.Service); ok {
			qs.QueryHandler(query{b})
			return
		}
	} else {
		if cs, ok := i.commandServices.get(b.Service); ok {
			cs.CommandHandler(command{b})
			return
		}
	}

	log.Info(log.Lbl("no_handler_or_callback"), b.IsResponse(), b.Id, b.Service, b.port)
}

func (i *Router) removeCallback(id uint32) {
	time.Sleep(time.Second)
	i.callbacks.delete(id)
}

func (i *Router) removeNetCallback(id uint32) {
	time.Sleep(time.Second)
	i.netcallbacks.delete(id)
}
