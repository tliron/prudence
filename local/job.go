package local

import (
	contextpkg "context"
	"sync/atomic"

	"github.com/tliron/prudence/platform"
)

//
// FuncJob
//

var nextKey uint64

type FuncJob struct {
	id uint64
	f  platform.JobFunc
}

func NewFuncJob(f platform.JobFunc) *FuncJob {
	return &FuncJob{
		id: atomic.AddUint64(&nextKey, 1),
		f:  f,
	}
}

// ([quartz.Job] interface)
func (self *FuncJob) Execute(context contextpkg.Context) {
	self.f()
}

// ([quartz.Job] interface)
func (self *FuncJob) Description() string {
	return "Prudence"
}

// ([quartz.Job] interface)
func (self *FuncJob) Key() int {
	return int(self.id)
}
