package auth

import (
	"container/heap"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCleanupQueueOrdersByExpiry(t *testing.T) {
	now := time.Now()

	q := &cleanupQueue{}
	heap.Init(q)

	heap.Push(q, cleanupItem{id: "later", expiresAt: now.Add(2 * time.Hour)})
	heap.Push(q, cleanupItem{id: "soonest", expiresAt: now.Add(1 * time.Minute)})
	heap.Push(q, cleanupItem{id: "middle", expiresAt: now.Add(1 * time.Hour)})

	first := heap.Pop(q).(cleanupItem)
	second := heap.Pop(q).(cleanupItem)
	third := heap.Pop(q).(cleanupItem)

	assert.Equal(t, "soonest", first.id)
	assert.Equal(t, "middle", second.id)
	assert.Equal(t, "later", third.id)
}
