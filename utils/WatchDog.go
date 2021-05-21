package utils

import (
	"github.com/tursom/GoCollections/collections"
	"sync"
	"sync/atomic"
	"time"
)

type WatchDog struct {
	feeding  bool
	callback func() bool
}

var watchDogMutex = sync.Mutex{}
var watchDogList = collections.NewLinkedList()
var watchDogId int32 = 0

func InitWatchDog(delay time.Duration) {
	currentWatchDogId := atomic.AddInt32(&watchDogId, 1)
	go func() {
		for currentWatchDogId == watchDogId {
			watchDogMutex.Lock()
			_ = collections.LoopMutable(watchDogList, func(element interface{}, iterator collections.MutableIterator) (err error) {
				watchDog := element.(*WatchDog)
				if !watchDog.feeding {
					if watchDog.callback() {
						_ = iterator.Remove()
					}
				}
				return
			})
			watchDogMutex.Unlock()
			time.Sleep(delay)
		}
	}()
}

func NewWatchDog(callback func() bool) *WatchDog {
	w := &WatchDog{true, callback}
	watchDogMutex.Lock()
	watchDogList.Add(w)
	watchDogMutex.Unlock()
	return w
}

func (w *WatchDog) Feed() {
	w.feeding = true
}
