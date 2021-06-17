package platform

import (
	"sync"
)

var startGroup *StartGroup
var startGroupLock sync.Mutex

func Start(startables []Startable) error {
	Stop()

	startGroupLock.Lock()
	defer startGroupLock.Unlock()

	startGroup = NewStartGroup(startables)
	startGroup.Start()

	return nil
}

func Stop() {
	startGroupLock.Lock()
	defer startGroupLock.Unlock()

	if startGroup != nil {
		startGroup.Stop()
		startGroup = nil
	}
}

//
// Startable
//

type Startable interface {
	Start() error
	Stop() error
}

//
// StartGroup
//

type StartGroup struct {
	Startables []Startable

	lock    sync.Mutex
	started sync.WaitGroup
}

type StartEntry struct {
	Startable Startable
}

func NewStartGroup(startables []Startable) *StartGroup {
	return &StartGroup{Startables: startables}
}

func (self *StartGroup) Start() {
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

func (self *StartGroup) Stop() {
	log.Info("stopping")

	self.lock.Lock()
	defer self.lock.Unlock()

	for _, startable := range self.Startables {
		if err := startable.Stop(); err != nil {
			log.Errorf("%s", err.Error())
		}
	}

	self.started.Wait()

	log.Info("stopped")
}
