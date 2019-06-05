package ipcrouter_test

import (
	"github.com/dist-ribut-us/ipcrouter"
	"github.com/dist-ribut-us/message"
	"github.com/dist-ribut-us/testutil"
	"github.com/dist-ribut-us/testutil/servicetest"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCommand(t *testing.T) {
	service, err := servicetest.Mock(12345)
	assert.NoError(t, err)

	sender, err := ipcrouter.New(testutil.GetPort())
	assert.NoError(t, err)

	go service.Run()
	defer service.Stop()

	msg := sender.
		Command(message.Test, "this is a test").
		To(service.GetPort()).
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
	s1, err := servicetest.Mock(6789)
	assert.NoError(t, err)
	s2, err := servicetest.Mock(1111)
	assert.NoError(t, err)

	go s1.Run()
	defer s1.Stop()
	go s2.Run()
	defer s2.Stop()

	s2.Router.
		Query(message.Test, "query").
		To(s1.GetPort()).
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
	sA, err := servicetest.Mock(111111)
	assert.NoError(t, err)
	sB, err := servicetest.MockWithRouter(222222, sA.Router)
	assert.NoError(t, err)
	sC, err := servicetest.Mock(333333)
	assert.NoError(t, err)

	assert.Equal(t, sA.GetPort(), sB.GetPort())

	go sA.Run()
	defer sA.Stop()
	go sC.Run()
	defer sC.Stop()

	msg := []byte{1, 2, 3, 4, 5}
	sA.Router.
		Command(message.Test, msg).
		To(sB.GetPort()).
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
		To(sC.GetPort()).
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
