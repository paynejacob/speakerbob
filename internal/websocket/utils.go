package websocket

type ConnectionSet map[uint64]*Connection

func (s ConnectionSet) Add(connection *Connection) {
	s[connection.Hash()] = connection
}

func (s ConnectionSet) Remove(connection *Connection) {
	delete(s, connection.Hash())
}

func (s ConnectionSet) Union(b ConnectionSet) ConnectionSet {
	newSet := ConnectionSet{}

	for hash, conn := range s {
		newSet[hash] = conn
	}

	for hash, conn := range b {
		newSet[hash] = conn
	}

	return newSet
}

func (s ConnectionSet) Values() []*Connection {
	connections := make([]*Connection, len(s))

	for _, conn := range s {
		connections = append(connections, conn)
	}

	return connections
}

type ChannelSet map[uint64]*Channel

func (s ChannelSet) Add(channel *Channel) {
	s[channel.Hash()] = channel
}

func (s ChannelSet) Remove(channel *Channel) {
	delete(s, channel.Hash())
}

func (s ChannelSet) Union(b ChannelSet) ChannelSet {
	newSet := ChannelSet{}

	for hash, conn := range s {
		newSet[hash] = conn
	}

	for hash, conn := range b {
		newSet[hash] = conn
	}

	return newSet
}

func (s ChannelSet) Values() []*Channel {
	channels := make([]*Channel, len(s))

	for _, conn := range s {
		channels = append(channels, conn)
	}

	return channels
}
