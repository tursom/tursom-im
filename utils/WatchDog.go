package utils

import (
	"fmt"
	"time"

	"github.com/tursom/GoCollections/collections"
	concurrent "github.com/tursom/GoCollections/concurrent/collections"
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
)

type WatchDog struct {
	lang.BaseObject
	feeding  int32
	life     int32
	node     collections.QueueNode[*WatchDog]
	callback func() bool
}

var watchDogList concurrent.ConcurrentLinkedQueue[*WatchDog]

func InitWatchDog() {
	go func() {
		for {
			start := time.Now().UnixNano()
			//fmt.Println("watch dog loop", watchDogList)
			_ = collections.LoopMutable[*WatchDog](&watchDogList, func(watchDog *WatchDog, iterator collections.MutableIterator[*WatchDog]) (err exceptions.Exception) {
				watchDog.feeding--
				if watchDog.feeding <= 0 {
					watchDog.feeding = watchDog.life
					_, err = exceptions.Try[any](func() (ret any, err exceptions.Exception) {
						if watchDog.callback() {
							_ = iterator.Remove()
						}
						return
					}, func(panic any) (ret any, err exceptions.Exception) {
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

func NewWatchDog(life int32, callback func() bool) *WatchDog {
	if life <= 0 {
		panic(exceptions.NewIllegalParameterException(fmt.Sprintf("watch dog feed lift must more than 0"), nil))
	}
	w := &WatchDog{feeding: life, life: life, callback: callback}
	node, err := watchDogList.OfferAndGetNode(w)
	if err != nil {
		err.PrintStackTrace()
		return nil
	}
	w.node = node
	return w
}

func (w *WatchDog) Feed() {
	w.feeding = w.life
}

func (w WatchDog) Life() int32 {
	return w.feeding
}

func (w *WatchDog) Kill() {
	exceptions.Exec0r0(w.node.Remove)
}
