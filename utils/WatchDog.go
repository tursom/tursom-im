package utils

import (
	"github.com/tursom/GoCollections/collections"
	"github.com/tursom/GoCollections/exceptions"
	"time"
)

type WatchDog struct {
	feeding  uint32
	life     uint32
	callback func() bool
}

var watchDogList = collections.NewConcurrentLinkedStack()

func InitWatchDog() {
	go func() {
		for {
			start := time.Now().UnixNano()
			//fmt.Println("watch dog loop", watchDogList)
			_ = collections.LoopMutable(watchDogList, func(element interface{}, iterator collections.MutableIterator) (err exceptions.Exception) {
				watchDog := element.(*WatchDog)
				watchDog.feeding--
				if watchDog.feeding == 0 {
					watchDog.feeding = watchDog.life
					_, _ = exceptions.Try(func() (ret interface{}, err exceptions.Exception) {
						if watchDog.callback() {
							_ = iterator.Remove()
						}
						return
					}, func(panic interface{}) (ret interface{}, err exceptions.Exception) {
						exceptions.PackageAny(panic).PrintStackTrace()
						_ = iterator.Remove()
						return
					})
				}
				return
			})
			end := time.Now().UnixNano()
			delay := time.Second - time.Nanosecond*time.Duration(end-start)
			if delay > 0 {
				time.Sleep(delay)
			}
		}
	}()
}

func NewWatchDog(life uint32, callback func() bool) *WatchDog {
	w := &WatchDog{life, life, callback}
	err := watchDogList.Push(w)
	if err != nil {
		err.PrintStackTrace()
		return nil
	}
	return w
}

func (w *WatchDog) Feed() {
	w.feeding = w.life
}

func (w WatchDog) Life() uint32 {
	return w.feeding
}