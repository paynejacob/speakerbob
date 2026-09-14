package websocket

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeBroadcaster struct {
	messages []interface{}
}

func (f *fakeBroadcaster) Broadcast(msg interface{}) {
	f.messages = append(f.messages, msg)
}

func TestBroadcastMessageDelegatesToConfiguredBroadcaster(t *testing.T) {
	fake := &fakeBroadcaster{}
	svc := &Service{Broadcaster: fake}

	conn := newFakeConn(svc, nil)
	svc.registerConnection(conn)

	require.Len(t, fake.messages, 2, "expected connection_count and presence messages")
	_, ok := fake.messages[0].(ConnectionCountMessage)
	assert.True(t, ok, "expected first delegated message to be ConnectionCountMessage")

	select {
	case <-conn.send:
		t.Fatal("in-process fan-out should not have run when a custom Broadcaster is configured")
	default:
	}
}

func TestBroadcastMessageDefaultsToInProcessFanOut(t *testing.T) {
	svc := &Service{}

	conn := newFakeConn(svc, nil)
	svc.registerConnection(conn)

	_, ok := (<-conn.send).(ConnectionCountMessage)
	assert.True(t, ok, "expected the default in-process broadcaster to deliver messages")
}
