package injector

import (
	"github.com/steromano87/uzi/v1/pkg/syntheticuser"
	"google.golang.org/grpc"
)

type Client struct {
	AgentClient
	syntheticuser.SpawnerClient
}

func NewRemoteClient(target string, opts ...grpc.DialOption) (Client, error) {
	clientConn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return Client{}, err
	}

	return newClient(clientConn), nil
}

func NewLocalClient(cc grpc.ClientConnInterface) Client {
	return newClient(cc)
}

func newClient(cc grpc.ClientConnInterface) Client {
	return Client{
		AgentClient:   NewAgentClient(cc),
		SpawnerClient: syntheticuser.NewSpawnerClient(cc),
	}
}
