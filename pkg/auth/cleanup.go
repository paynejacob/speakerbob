package auth

import "time"

type cleanupItem struct {
	id        string
	expiresAt time.Time
}

// cleanupQueue is a container/heap.Interface ordering pending token
// deletions by expiry, so Service.Run only wakes when something is actually
// due instead of periodically listing every token.
type cleanupQueue []cleanupItem

func (q cleanupQueue) Len() int           { return len(q) }
func (q cleanupQueue) Less(i, j int) bool { return q[i].expiresAt.Before(q[j].expiresAt) }
func (q cleanupQueue) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }

func (q *cleanupQueue) Push(x interface{}) {
	*q = append(*q, x.(cleanupItem))
}

func (q *cleanupQueue) Pop() interface{} {
	old := *q
	n := len(old)
	item := old[n-1]
	*q = old[:n-1]
	return item
}
