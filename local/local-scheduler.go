package local

import (
	contextpkg "context"
	"sync"

	"github.com/reugn/go-quartz/quartz"
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/commonlog/sink"
	"github.com/tliron/go-ard"
	"github.com/tliron/prudence/platform"
)

func init() {
	sink.SetDefaultQuartzLogger("prudence", "local", "scheduler")
}

//
// LocalScheduler
//

type LocalScheduler struct {
	scheduler quartz.Scheduler
	queue     []func() error
	queueLock sync.Mutex
}

func NewLocalScheduler() *LocalScheduler {
	return &LocalScheduler{
		scheduler: quartz.NewStdScheduler(),
	}
}

// ([platform.CreateFunc] signature)
func CreateLocalScheduler(jsContext *commonjs.Context, config ard.StringMap) (any, error) {
	return NewLocalScheduler(), nil
}

// ([platform.Scheduler] interface)
func (self *LocalScheduler) Schedule(cronPattern string, job platform.JobFunc) error {
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

// ([platform.Startable] interface)
func (self *LocalScheduler) Start() error {
	self.scheduler.Start(contextpkg.TODO())
	log.Info("started Quartz scheduler")

	self.queueLock.Lock()
	defer self.queueLock.Unlock()
	for _, f := range self.queue {
		if err := f(); err != nil {
			log.Error(err.Error())
		}
	}
	self.queue = nil

	return nil
}

// ([platform.Startable] interface)
func (self *LocalScheduler) Stop(stopContext contextpkg.Context) error {
	self.scheduler.Stop()
	log.Info("stopped Quartz scheduler")
	return nil
}

func (self *LocalScheduler) schedule(cronPattern string, job platform.JobFunc) error {
	log.Infof("scheduling job at: %s", cronPattern)

	if trigger, err := quartz.NewCronTrigger(cronPattern); err == nil {
		return self.scheduler.ScheduleJob(contextpkg.TODO(), NewFuncJob(job), trigger)
	} else {
		return err
	}
}
