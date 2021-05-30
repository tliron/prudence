package platform

import "sync"

var stopChannel = make(chan bool)

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
	Entries []StartEntry

	lock sync.Mutex
}

type StartEntry struct {
	Startable Startable
	Stopped   chan bool
}

func NewStartGroup(startables []Startable) *StartGroup {
	entries := make([]StartEntry, len(startables))
	for index, startable := range startables {
		entries[index] = StartEntry{
			Startable: startable,
			Stopped:   make(chan bool),
		}
	}
	return &StartGroup{Entries: entries}
}

func (self *StartGroup) Start() {
	self.lock.Lock()
	defer self.lock.Unlock()

	for _, entry := range self.Entries {
		entry_ := entry // closure capture
		go func() {
			if err := entry_.Startable.Start(); err != nil {
				log.Errorf("%s", err.Error())
			}
			entry_.Stopped <- true
		}()
	}
}

func (self *StartGroup) Stop() {
	self.lock.Lock()
	defer self.lock.Unlock()

	for _, entry := range self.Entries {
		if err := entry.Startable.Stop(); err != nil {
			log.Errorf("%s", err.Error())
		}
		<-entry.Stopped
	}
}
