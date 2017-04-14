package ipcrouter

import (
	"github.com/dist-ribut-us/log"
	"github.com/dist-ribut-us/message"
	"github.com/dist-ribut-us/rnet"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var portInc uint16 = 5555

func getPort() rnet.Port {
	portInc++
	return rnet.Port(portInc)
}

func init() {
	log.Mute()
}

func TestHandler(t *testing.T) {
	r1, err := New(getPort())
	assert.NoError(t, err)
	r2, err := New(getPort())
	assert.NoError(t, err)
	var service uint32 = 123

	ch := make(chan *message.Header)
	r1.Register(service, func(b *Base) {
		ch <- b.Header
	})
	go r1.Run()

	msg := r2.
		Base(message.Test, "this is a test").
		To(r1.Port()).
		SetService(service)
	id := msg.Id
	msg.Send(nil)

	select {
	case out := <-ch:
		assert.Equal(t, "this is a test", out.BodyString())
		assert.Equal(t, message.Test, out.GetType())
		assert.Equal(t, id, out.Id)
	case <-time.After(time.Millisecond * 20):
		t.Error("timed out")
	}
}

func TestResponseCallback(t *testing.T) {
	r1, err := New(getPort())
	assert.NoError(t, err)
	r2, err := New(getPort())
	assert.NoError(t, err)
	var service uint32 = 123

	ch := make(chan string)
	r1.Register(service, func(b *Base) {
		assert.True(t, b.IsQuery())
		ch <- b.BodyString()
		b.Respond("response")
	})
	go r1.Run()
	go r2.Run()

	log.Info("r1", r1.Port())
	log.Info("r2", r2.Port())

	r2.
		Query(message.Test, "query").
		To(r1.Port()).
		SetService(service).
		Send(func(b *Base) {
			assert.True(t, b.IsResponse())
			ch <- b.BodyString()
		})

	select {
	case out := <-ch:
		assert.Equal(t, "query", out)
	case <-time.After(time.Millisecond * 10):
		t.Error("timed out")
	}

	select {
	case out := <-ch:
		assert.Equal(t, "response", out)
	case <-time.After(time.Millisecond * 10):
		t.Error("timed out")
	}
}

func TestSamePort(t *testing.T) {
	p1, err := New(getPort())
	assert.NoError(t, err)
	p2, err := New(getPort())
	assert.NoError(t, err)

	go p1.Run()
	go p2.Run()

	var serviceAID uint32 = 111111
	var serviceBID uint32 = 222222
	var serviceCID uint32 = 333333

	out := make(chan []byte)
	fn := func(b *Base) {
		out <- b.Body
	}
	p1.Register(serviceAID, fn)
	p1.Register(serviceBID, fn)
	p2.Register(serviceCID, fn)

	msg := []byte{1, 2, 3, 4, 5}
	p1.
		Base(message.Test, msg).
		To(p1.Port()).
		SetService(serviceBID).
		Send(nil)

	select {
	case got := <-out:
		msg[0] = 111 // assert that the reference is equal
		assert.Equal(t, msg, got)
	case <-time.After(time.Millisecond * 10):
		t.Error("timeout")
	}

	p1.
		Base(message.Test, msg).
		To(p2.Port()).
		SetService(serviceCID).
		Send(nil)

	select {
	case got := <-out:
		assert.Equal(t, msg, got)
		msg[0] = 222
		assert.NotEqual(t, msg, got)
	case <-time.After(time.Millisecond * 10):
		t.Error("timeout")
	}
}
