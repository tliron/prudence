package platform

import (
	"context"
	contextpkg "context"
	"sync"
	"time"
)

var startableGroup *StartableGroup
var startableGroupLock sync.Mutex

func Start(startables []Startable, stopTimeout time.Duration) error {
	Stop()

	startableGroupLock.Lock()
	defer startableGroupLock.Unlock()

	startableGroup = NewStartableGroup(startables, stopTimeout)
	startableGroup.Start()

	return nil
}

func Stop() {
	startableGroupLock.Lock()
	defer startableGroupLock.Unlock()

	if startableGroup != nil {
		startableGroup.Stop()
		startableGroup = nil
	}
}

//
// Startable
//

type Startable interface {
	Start() error
	Stop(stopContext contextpkg.Context) error
}

//
// StartableGroup
//

type StartableGroup struct {
	Startables  []Startable
	StopTimeout time.Duration

	lock    sync.Mutex
	started sync.WaitGroup
}

type StartEntry struct {
	Startable Startable
}

func NewStartableGroup(startables []Startable, stopTimeout time.Duration) *StartableGroup {
	return &StartableGroup{
		Startables:  startables,
		StopTimeout: stopTimeout,
	}
}

func (self *StartableGroup) Start() {
	log.Info("starting")

	self.lock.Lock()
	defer self.lock.Unlock()

	for _, startable := range self.Startables {
		go func(startable Startable) {
			self.started.Add(1)
			defer self.started.Done()

			if err := startable.Start(); err != nil {
				log.Errorf("%s", err.Error())
			}
		}(startable)
	}
}

func (self *StartableGroup) Stop() {
	log.Info("stopping")

	self.lock.Lock()
	defer self.lock.Unlock()

	// Note: we are not using errgroup because we want to catch all errors,
	// not just the first one

	stopContext, cancel := contextpkg.WithTimeout(context.Background(), self.StopTimeout)

	for i := len(self.Startables) - 1; i >= 0; i-- {
		if err := self.Startables[i].Stop(stopContext); err != nil {
			log.Errorf("%s", err.Error())
		}
	}

	self.started.Wait()

	cancel()

	log.Info("stopped")
}
