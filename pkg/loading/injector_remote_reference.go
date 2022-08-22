package loading

import "github.com/steromano87/harkonnen/v1/pkg/messaging"

type InjectorRemoteReference struct {
	Name      string
	Weight    int
	Messenger messaging.Messenger
}
