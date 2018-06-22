package testservice

import (
	"github.com/dist-ribut-us/ipcrouter"
	"github.com/dist-ribut-us/rnet"
)

// TestService provides hooks for simulating a service with a router.
type TestService struct {
	*Chan
	ID uint32
	*ipcrouter.Router
}

type Chan struct {
	Query       chan ipcrouter.Query
	Command     chan ipcrouter.Command
	Response    chan ipcrouter.Response
	NetQuery    chan ipcrouter.NetQuery
	NetCommand  chan ipcrouter.NetCommand
	NetResponse chan ipcrouter.NetResponse
}

func New(id uint32, port rnet.Port) (*TestService, error) {
	r, err := ipcrouter.New(port)
	if err != nil {
		return nil, err
	}
	return NewWithRouter(id, r)
}

func NewWithRouter(id uint32, router *ipcrouter.Router) (*TestService, error) {
	ts := &TestService{
		Chan: &Chan{
			Query:       make(chan ipcrouter.Query),
			Command:     make(chan ipcrouter.Command),
			Response:    make(chan ipcrouter.Response),
			NetQuery:    make(chan ipcrouter.NetQuery),
			NetCommand:  make(chan ipcrouter.NetCommand),
			NetResponse: make(chan ipcrouter.NetResponse),
		},
		ID:     id,
		Router: router,
	}
	return ts, router.Register(ts)
}

func (ts *TestService) ServiceID() uint32 {
	return ts.ID
}

func (ts *TestService) CommandHandler(cmd ipcrouter.Command) {
	ts.Chan.Command <- cmd
}

func (ts *TestService) QueryHandler(q ipcrouter.Query) {
	ts.Chan.Query <- q
}

func (ts *TestService) Responder(r ipcrouter.Response) {
	ts.Chan.Response <- r
}

func (ts *TestService) NetQueryHandler(nq ipcrouter.NetQuery) {
	ts.Chan.NetQuery <- nq
}

func (ts *TestService) NetCommandHandler(nc ipcrouter.NetCommand) {
	ts.Chan.NetCommand <- nc
}

func (ts *TestService) NetResponder(r ipcrouter.NetResponse) {
	ts.Chan.NetResponse <- r
}
