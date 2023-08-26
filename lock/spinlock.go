package lock

import (
	"runtime"
	"sync/atomic"
)

type SpinLock int32

const maxWaitTimes = 1 << 5

func (s *SpinLock) Lock() {
	waitTimes := 1
	for !atomic.CompareAndSwapInt32((*int32)(s), 0, 1) {
		for i := 0; i < waitTimes; i++ {
			runtime.Gosched()
		}
		if waitTimes >= maxWaitTimes {
			waitTimes = 1
		}
		waitTimes = waitTimes << 1
	}
}

func (s *SpinLock) Unlock() {
	atomic.StoreInt32((*int32)(s), 0)
}

func NewSpinLock() *SpinLock {
	return new(SpinLock)
}
