package websocket

// Broadcaster fans a message out to every connected client. The default
// implementation does this in-process; a future implementation backed by a
// pub-sub system (e.g. Redis) could satisfy the same interface to let a
// message broadcast from one pod reach clients connected to another.
type Broadcaster interface {
	Broadcast(msg interface{})
}

type inProcessBroadcaster struct {
	service *Service
}

func (b *inProcessBroadcaster) Broadcast(msg interface{}) {
	b.service.m.RLock()
	defer b.service.m.RUnlock()

	for i := range b.service.connections {
		b.service.connections[i].SendMessage(msg)
	}
}
