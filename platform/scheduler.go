package platform

//
// Scheduler
//

var scheduler Scheduler

type Scheduler interface {
	Schedule(cronPattern string, job func()) error
}

func SetScheduler(scheduler_ Scheduler) {
	scheduler = scheduler_
}

func GetScheduler() Scheduler {
	return scheduler
}
