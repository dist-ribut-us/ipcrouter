package ipcroutertest

import (
	"github.com/dist-ribut-us/ipcrouter"
	"github.com/dist-ribut-us/ipcrouter/testservice"
	"github.com/dist-ribut-us/log"
	"github.com/dist-ribut-us/message"
	"github.com/dist-ribut-us/rnet"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var getPort = rnet.NewPortIncrementer(5555)

func init() {
	log.Mute()
}

func TestCommand(t *testing.T) {
	service, err := testservice.New(12345, getPort.Next())
	assert.NoError(t, err)

	sender, err := ipcrouter.New(getPort.Next())
	assert.NoError(t, err)

	go service.Run()
	defer service.Stop()

	msg := sender.
		Command(message.Test, "this is a test").
		To(service.Port()).
		SetService(service.ServiceID())
	id := msg.GetId()
	msg.Send(nil)

	select {
	case cmd := <-service.Chan.Command:
		assert.Equal(t, "this is a test", cmd.BodyString())
		assert.Equal(t, message.Test, cmd.GetType())
		assert.Equal(t, id, cmd.GetId())
	case <-time.After(time.Millisecond * 20):
		t.Error("timed out")
	}
}

func TestQueryAndCallback(t *testing.T) {
	s1, err := testservice.New(6789, getPort.Next())
	assert.NoError(t, err)
	s2, err := testservice.New(1111, getPort.Next())
	assert.NoError(t, err)

	go s1.Run()
	defer s1.Stop()
	go s2.Run()
	defer s2.Stop()

	s2.Router.
		Query(message.Test, "query").
		To(s1.Port()).
		SetService(s1.ServiceID()).
		Send(s2.Responder)

	select {
	case q := <-s1.Chan.Query:
		assert.Equal(t, "query", q.BodyString())
		q.Respond("response")
	case <-time.After(time.Millisecond * 10):
		t.Error("timed out")
	}

	select {
	case r := <-s2.Chan.Response:
		assert.Equal(t, "response", r.BodyString())
	case <-time.After(time.Millisecond * 10):
		t.Error("timed out")
	}
}

func TestSamePort(t *testing.T) {
	sA, err := testservice.New(111111, getPort.Next())
	assert.NoError(t, err)
	sB, err := testservice.NewWithRouter(222222, sA.Router)
	assert.NoError(t, err)
	sC, err := testservice.New(333333, getPort.Next())
	assert.NoError(t, err)

	assert.Equal(t, sA.Port(), sB.Port())

	go sA.Run()
	defer sA.Stop()
	go sC.Run()
	defer sC.Stop()

	msg := []byte{1, 2, 3, 4, 5}
	sA.Router.
		Command(message.Test, msg).
		To(sB.Port()).
		SetService(sB.ServiceID()).
		Send(nil)

	select {
	case cmd := <-sB.Chan.Command:
		msg[0] = 111 // by chaning a byte in msg we check assert that the reference is equal
		assert.Equal(t, msg, cmd.GetBody())
	case <-time.After(time.Millisecond * 10):
		t.Error("timeout")
	}

	sA.Router.
		Command(message.Test, msg).
		To(sC.Port()).
		SetService(sC.ServiceID()).
		Send(nil)

	select {
	case cmd := <-sC.Chan.Command:
		assert.Equal(t, msg, cmd.GetBody())
		msg[0] = 222
		assert.NotEqual(t, msg, cmd.GetBody())
	case <-time.After(time.Millisecond * 10):
		t.Error("timeout")
	}
}
