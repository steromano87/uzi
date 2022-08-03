package loading

import "github.com/steromano87/harkonnen/v1/pkg/network"

type InjectorRemoteReference struct {
	Name      string
	Weight    int
	Messenger network.Messenger
}
