package injector

import "github.com/steromano87/harkonnen/v1/pkg/messaging"

type RemoteReference struct {
	ID     string
	Weight int
	messaging.Messenger
}
