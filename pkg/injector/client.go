package injector

import (
	"github.com/steromano87/harkonnen/v1/pkg/syntheticuser"
	"github.com/steromano87/harkonnen/v1/pkg/workspace"
	"google.golang.org/grpc"
)

type Client struct {
	AgentClient
	syntheticuser.SpawnerClient
	workspace.WorkspaceClient
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
		AgentClient:     NewAgentClient(cc),
		SpawnerClient:   syntheticuser.NewSpawnerClient(cc),
		WorkspaceClient: workspace.NewWorkspaceClient(cc),
	}
}
