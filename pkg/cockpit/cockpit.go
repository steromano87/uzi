package cockpit

import (
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/steromano87/harkonnen/v1/pkg/scheduler"
)

type Cockpit struct {
	ctx Context

	localInjector      *injector.Injector
	injectorReferences map[string]*injector.Reference

	scheduler   scheduler.Scheduler
	loadProfile scheduler.LoadProfile
}

func New(ctx Context, loadProfile scheduler.LoadProfile) *Cockpit {
	cockpit := new(Cockpit)
	cockpit.ctx = ctx
	cockpit.injectorReferences = make(map[string]*injector.Reference)
	cockpit.loadProfile = loadProfile
	cockpit.scheduler = scheduler.NewFixedIntervalScheduler(
		cockpit.loadProfile,
		cockpit.injectorReferences,
		cockpit.ctx.Config().GetDuration("cockpit.scheduler.updateInterval"))

	return cockpit
}

func (c *Cockpit) AddInjector(ID string, weight int, messenger messaging.Messenger) {
	c.injectorReferences[ID] = &injector.Reference{
		Weight:    weight,
		Messenger: messenger,
	}
}

func (c *Cockpit) Start() {

}
