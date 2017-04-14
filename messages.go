package ipcrouter

import (
	"github.com/dist-ribut-us/errors"
	"github.com/dist-ribut-us/ipc"
	"github.com/dist-ribut-us/log"
	"github.com/dist-ribut-us/message"
	"github.com/dist-ribut-us/rnet"
)

// ErrTypesDoNotMatch is thrown when trying to convert a Type to the wrong
// type
const ErrTypesDoNotMatch = errors.String("Types do not match")

// ToBase a message body to get it's type
func (i *Router) toBase(m *ipc.Package) *Base {
	h := message.Unmarshal(m.Body)
	if h == nil {
		return nil
	}
	h.Id = m.ID
	return &Base{
		Header: h,
		proc:   i,
		port:   m.Addr.Port(),
	}
}

// Base provides a base message type. It wraps message.Header and provides
// helper functions for simple query and response messages.
type Base struct {
	*message.Header
	proc *Router
	port rnet.Port
}

// To sets the port Send will send to.
func (b *Base) To(port rnet.Port) *Base {
	b.port = port
	return b
}

// ToNet sets the fields for a message to be sent to the Overlay service and
// then out over the net.
func (b *Base) ToNet(overlayPort rnet.Port, netAddr *rnet.Addr, remoteServiceID uint32) *Base {
	return b.
		SetFlag(message.ToNet).
		SetAddr(netAddr).
		SetService(remoteServiceID).
		To(overlayPort)
}

// SetAddr sets the address on a message. This indicates the network address
// where the message should be sent, most likely by overlay.
func (b *Base) SetAddr(addr *rnet.Addr) *Base {
	b.Header.SetAddr(addr)
	return b
}

// SetFlag field on the Header
func (b *Base) SetFlag(flag message.BitFlag) *Base {
	b.Header.SetFlag(flag)
	return b
}

// SetService field on the header
func (b *Base) SetService(service uint32) *Base {
	b.Header.Service = service
	return b
}

// Port returns the base port - this is the ipc port that the message came from
// or that it is sent to send to.
func (b *Base) Port() rnet.Port {
	return b.port
}

// Respond to a query
func (b *Base) Respond(body interface{}) {
	r := &Base{
		Header: &message.Header{
			Id:     b.Id,
			Type32: b.Type32,
			Flags:  uint32(message.ResponseFlag),
		},
		proc: b.proc,
		port: b.port,
	}
	r.SetBody(body)

	log.Info(b.IsFromNet(), b.Addrpb)

	if b.IsFromNet() {
		r.SetFlag(message.ToNet)
		r.Addrpb = b.Addrpb
	}

	r.Send(nil)
}

// Query creates a basic query.
func (i *Router) Query(t message.Type, body interface{}) *Base {
	h := message.NewHeader(t, body)
	h.SetFlag(message.QueryFlag)
	return &Base{
		Header: h,
		proc:   i,
	}
}

// Base creates a basic message with no flags. Body can be either a proto
// message or a byte slice.
func (i *Router) Base(t message.Type, body interface{}) *Base {
	return &Base{
		Header: message.NewHeader(t, body),
		proc:   i,
	}
}

// Send a message. If callback is not nil, the reponse will be sent to the
// callback
func (b *Base) Send(callback Handler) {
	id := b.Id

	if callback != nil {
		b.proc.callbacks.set(id, callback)
		go b.proc.removeCallback(id)
	}

	if b.port == b.proc.Port() {
		go b.proc.baseHandler(b)
		return
	}

	b.Id = 0
	b.proc.ipc.Send(id, b.Marshal(), b.Port())
}

// RequestServicePort is a shorthand to request a service port from pool.
func (i *Router) RequestServicePort(serviceName string, pool rnet.Port, callback Handler) {
	i.
		Query(message.GetPort, serviceName).
		To(pool).
		SetService(message.PoolService).
		Send(callback)
}

// RegisterWithOverlay is a shorthand to register a service with overlay.
func (i *Router) RegisterWithOverlay(serviceID uint32, overlay rnet.Port) {
	i.
		Base(message.RegisterService, serviceID).
		To(overlay).
		SetService(message.OverlayService).
		Send(nil)
}
