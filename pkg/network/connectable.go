package network

type Connectable interface {
	GetConnection() *Connection
}

type Connection struct {
	Network *Network
	IPv4    *string
}

type Connections = map[string]*Connection

func NewConnections(networks ...Connectable) Connections {
	connections := make(Connections)
	for _, network := range networks {
		conn := network.GetConnection()
		connections[conn.Network.Name] = conn
	}

	return connections
}
