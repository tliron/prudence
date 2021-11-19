package local

import (
	"sync/atomic"
)

//
// FuncJob
//

var nextKey int64

type FuncJob struct {
	key int
	f   func()
}

func NewFuncJob(f func()) *FuncJob {
	return &FuncJob{
		key: int(atomic.AddInt64(&nextKey, 1)),
		f:   f,
	}
}

// quartz.Job interface
func (self *FuncJob) Execute() {
	self.f()
}

// quartz.Job interface
func (self *FuncJob) Description() string {
	return "Prudence"
}

// quartz.Job interface
func (self *FuncJob) Key() int {
	return self.key
}
