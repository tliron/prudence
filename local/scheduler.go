package local

import (
	contextpkg "context"
	"sync"

	"github.com/reugn/go-quartz/quartz"
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterType("LocalScheduler", CreateLocalScheduler)
}

//
// LocalScheduler
//

type LocalScheduler struct {
	scheduler *quartz.StdScheduler
	queue     []func() error
	queueLock sync.Mutex
}

func NewLocalScheduler() *LocalScheduler {
	return &LocalScheduler{
		scheduler: quartz.NewStdScheduler(),
	}
}

// platform.CreateFunc signature
func CreateLocalScheduler(config ard.StringMap, context *js.Context) (interface{}, error) {
	return NewLocalScheduler(), nil
}

// platform.Scheduler interface
func (self *LocalScheduler) Schedule(cronPattern string, job func()) error {
	if self.scheduler.IsStarted() {
		return self.schedule(cronPattern, job)
	} else {
		// Call after scheduler is started
		self.queueLock.Lock()
		defer self.queueLock.Unlock()
		self.queue = append(self.queue, func() error {
			return self.schedule(cronPattern, job)
		})
		return nil
	}
}

// platform.Startable interface
func (self *LocalScheduler) Start() error {
	self.scheduler.Start()
	log.Info("started Quartz scheduler")

	self.queueLock.Lock()
	defer self.queueLock.Unlock()
	for _, f := range self.queue {
		if err := f(); err != nil {
			log.Errorf("%s", err)
		}
	}
	self.queue = nil

	return nil
}

// platform.Startable interface
func (self *LocalScheduler) Stop(stopContext contextpkg.Context) error {
	self.scheduler.Stop()
	log.Info("stopped Quartz scheduler")
	return nil
}

func (self *LocalScheduler) schedule(cronPattern string, job func()) error {
	log.Infof("scheduling task at: %s", cronPattern)

	if trigger, err := quartz.NewCronTrigger(cronPattern); err == nil {
		return self.scheduler.ScheduleJob(NewFuncJob(job), trigger)
	} else {
		return err
	}
}
