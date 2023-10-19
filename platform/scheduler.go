package platform

//
// Scheduler
//

type JobFunc func()

var scheduler Scheduler

type Scheduler interface {
	Schedule(cronPattern string, job JobFunc) error
}

func SetScheduler(scheduler_ Scheduler) {
	scheduler = scheduler_
}

func GetScheduler() Scheduler {
	return scheduler
}
