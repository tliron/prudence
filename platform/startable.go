package platform

import (
	"sync"
)

var startGroup *StartGroup

func Start(startables []Startable) error {
	Stop()

	log.Info("starting")

	startGroup = NewStartGroup(startables)
	startGroup.Start()

	// Block forever
	<-make(chan bool, 0)

	return nil
}

func Stop() {
	if startGroup != nil {
		log.Info("stopping")
		startGroup.Stop()
	}
}

func Restart() {
	if startGroup != nil {
		log.Info("restarting")
		startGroup.Stop()
		startGroup.Start()
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
	self.lock.Lock()
	defer self.lock.Unlock()

	for _, startable := range self.Startables {
		startable_ := startable // closure capture
		self.started.Add(1)
		go func() {
			if err := startable_.Start(); err != nil {
				log.Errorf("%s", err.Error())
			}
			self.started.Done()
		}()
	}
}

func (self *StartGroup) Stop() {
	self.lock.Lock()
	defer self.lock.Unlock()

	for _, startable := range self.Startables {
		if err := startable.Stop(); err != nil {
			log.Errorf("%s", err.Error())
		}
	}

	self.started.Wait()
}
