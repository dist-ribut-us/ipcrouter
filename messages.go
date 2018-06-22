package ipcrouter

import (
	"github.com/dist-ribut-us/errors"
	"github.com/dist-ribut-us/ipc"
	"github.com/dist-ribut-us/log"
	"github.com/dist-ribut-us/message"
	"github.com/dist-ribut-us/overlay/overlaymessages"
	"github.com/dist-ribut-us/rnet"
	"github.com/golang/protobuf/proto"
)

// ErrTypesDoNotMatch is thrown when trying to convert a Type to the wrong
// type
const ErrTypesDoNotMatch = errors.String("Types do not match")

// ToBase a message body to get it's type
func (i *Router) toBase(m *ipc.Package) *base {
	h := message.Unmarshal(m.Body)
	if h == nil {
		return nil
	}
	h.Id = m.ID
	return &base{
		Header: h,
		proc:   i,
		port:   m.Addr.Port(),
	}
}

type base struct {
	*message.Header
	proc *Router
	port rnet.Port
}

type baseIfc interface {
	GetType32() uint32
	GetType() message.Type
	GetService() uint32
	GetFlags() uint32
	GetBody() []byte
	GetAddr() *rnet.Addr
	BodyString() string
	BodyToUint32() uint32
	Unmarshal(proto.Message) error
	GetId() uint32
	Port() rnet.Port
}

// Sender is a message as it's being built, it can be sent.
type Sender interface {
	baseIfc
	Send(ResponseCallback)
	SetService(service uint32) Sender
	SetAddr(addr *rnet.Addr) Sender
	SetFlag(flag message.BitFlag) Sender
	To(port rnet.Port) Sender
	SendToNet(netAddr *rnet.Addr, callback NetResponseCallback)
}

// Query is message that expects a response
type Query interface {
	baseIfc
	Respond(body interface{})
	isQuery()
}
type query struct{ *base }

func (query) isQuery() {}

// Response is handled by a query callback. Responses not handled will
// fallthrough as Commands.
type Response interface {
	baseIfc
	isResponse()
}
type response struct{ *base }

func (response) isResponse() {}

// Command is a message that does not expect a response.
type Command interface {
	baseIfc
	isCommand()
}
type command struct{ *base }

func (command) isCommand() {}

// baseNetMsg is shared by all net commands
type baseNetMsg interface {
	baseIfc
	GetNodeID() []byte
	GetAddrpb() *message.Addrpb
	GetRouterPort() rnet.Port
	GetHeader() *message.Header
}

// NetSendRequest is sent by one local service to another local service that
// can handle sending requests to the internet. Generally that service will be
// the Overlay service.
type NetSendRequest interface {
	baseNetMsg
	isNetSendRequest()
}
type netSendRequest struct{ *base }

func (netSendRequest) isNetSendRequest() {}

// NetCommand is message sent over the network that does not expect a response
type NetCommand interface {
	baseNetMsg
	isNetCommand()
}
type netCommand struct{ *base }

func (netCommand) isNetCommand() {}

// NetQuery is a message sent over the network that expects a response
type NetQuery interface {
	baseNetMsg
	Respond(body interface{})
	isNetQuery()
}
type netQuery struct{ *base }

func (netQuery) isNetQuery() {}

// NetResponse is a response to a NetQuery
type NetResponse interface {
	baseNetMsg
	isNetResponse()
}
type netResponse struct{ *base }

func (netResponse) isNetResponse() {}

// Service can be registered with a Router.
type Service interface {
	ServiceID() uint32
}

// QueryService will handle Queries sent to it
type QueryService interface {
	Service
	QueryHandler(Query)
}

// CommandService will handle Queries sent to it
type CommandService interface {
	Service
	CommandHandler(Command)
}

// NetSenderService will handle Queries sent to it
type NetSenderService interface {
	Service
	NetSend(NetSendRequest)
}

// NetQueryService will handle Queries sent to it
type NetQueryService interface {
	Service
	NetQueryHandler(NetQuery)
}

// NetCommandService will handle Queries sent to it
type NetCommandService interface {
	Service
	NetCommandHandler(NetCommand)
}

// ResponseCallback can be registered when sending a Query
type ResponseCallback func(Response)

// NetResponseCallback can be registered when sending a NetQuery
type NetResponseCallback func(NetResponse)

// To sets the port Send will send to.
func (b *base) To(port rnet.Port) Sender {
	b.port = port
	return b
}

// SetAddr sets the address on a message. This indicates the network address
// where the message should be sent, most likely by overlay.
func (b *base) SetAddr(addr *rnet.Addr) Sender {
	b.Header.SetAddr(addr)
	return b
}

func (b *base) GetRouterPort() rnet.Port {
	return b.proc.Port()
}

func (b *base) GetHeader() *message.Header {
	return b.Header
}

// SetFlag field on the Header
func (b *base) SetFlag(flag message.BitFlag) Sender {
	b.Header.SetFlag(flag)
	return b
}

// SetService field on the header
func (b *base) SetService(service uint32) Sender {
	b.Header.Service = service
	return b
}

// Port returns the base port - this is the ipc port that the message came from
// or that it is sent to send to.
func (b *base) Port() rnet.Port {
	return b.port
}

// Respond to a query
func (b *base) Respond(body interface{}) {
	r := &base{
		Header: &message.Header{
			Id:     b.Id,
			Type32: b.Type32,
			Flags:  uint32(message.ResponseFlag),
		},
		proc: b.proc,
		port: b.port,
	}
	r.SetBody(body)

	log.Debug(b.IsFromNet(), b.Addrpb)

	if b.IsFromNet() {
		r.SetFlag(message.ToNet)
		r.Addrpb = b.Addrpb
	}

	r.Send(nil)
}

// Query creates a basic query.
func (i *Router) Query(t message.Type, body interface{}) Sender {
	h := message.NewHeader(t, body)
	h.SetFlag(message.QueryFlag)
	return &base{
		Header: h,
		proc:   i,
	}
}

// Command creates a basic message with no flags. Body can be either a proto
// message or a byte slice.
func (i *Router) Command(t message.Type, body interface{}) Sender {
	return &base{
		Header: message.NewHeader(t, body),
		proc:   i,
	}
}

// Send a message. If callback is not nil, the reponse will be sent to the
// callback
func (b *base) Send(callback ResponseCallback) {
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

func (b *base) SendToNet(netAddr *rnet.Addr, callback NetResponseCallback) {
	b.
		SetFlag(message.ToNet).
		SetAddr(netAddr).
		To(b.proc.NetSenderPort)

	id := b.Id
	if callback != nil {
		b.proc.netcallbacks.set(id, callback)
		go b.proc.removeNetCallback(id)
	}

	b.Id = 0
	b.proc.ipc.Send(id, b.Marshal(), b.Port())
}

// RequestServicePort is a shorthand to request a service port from pool.
func (i *Router) RequestServicePort(serviceName string, pool rnet.Port, callback ResponseCallback) {
	i.
		Query(message.GetPort, serviceName).
		To(pool).
		SetService(message.PoolService).
		Send(callback)
}

// RegisterWithOverlay is a shorthand to register a service with overlay.
func (i *Router) RegisterWithOverlay(serviceID uint32, overlay rnet.Port) {
	i.
		Command(message.RegisterService, serviceID).
		To(overlay).
		SetService(overlaymessages.ServiceID).
		Send(nil)
}
