package injector

import "github.com/steromano87/harkonnen/v1/pkg/messaging"

type RemoteReference struct {
	Weight int
	messaging.Messenger
}
