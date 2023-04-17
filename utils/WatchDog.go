package utils

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tursom/GoCollections/collections"
	concurrent "github.com/tursom/GoCollections/concurrent/collections"
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
)

type WatchDog struct {
	lang.BaseObject
	feeding  int32
	life     int32
	callback func()
}

var (
	watchDogList concurrent.ConcurrentLinkedQueue[*WatchDog]
)

// init watch dog loop handler
func init() {
	go watchDogLooper()
}

func watchDogLooper() {
	log.Info("utils/WatchDog.go: init watch dog looper")

	tick := time.Tick(time.Second)

	for {
		<-tick

		dead := 0

		// call collections.LoopMutable to loop watchDogList
		if err := collections.LoopMutable[*WatchDog](&watchDogList, func(watchDog *WatchDog, iterator collections.MutableIterator[*WatchDog]) (err exceptions.Exception) {
			watchDog.feeding--

			if watchDog.feeding <= 0 {
				_, err = exceptions.Try[any](func() (ret any, err exceptions.Exception) {
					defer func() {
						_ = iterator.Remove()
						dead++
					}()

					if watchDog.callback != nil {
						watchDog.callback()
					}
					return
				}, func(panic any) (ret any, err exceptions.Exception) {
					return nil, exceptions.PackageAny(panic)
				})
			}
			return
		}); err != nil {
			log.WithFields(log.Fields{
				"err":        err,
				"stackTrace": exceptions.GetStackTraceString(err),
			}).Warn("utils/WatchDog.go: loop watch dog err")
		}

		log.WithFields(log.Fields{
			"dead": dead,
			"live": watchDogList.Size(),
		}).Info("utils/WatchDog.go: loop finished")
	}
}

// NewWatchDog get new WatchDog with life cycle and callback
func NewWatchDog(life int32, callback func()) *WatchDog {
	if life <= 0 {
		panic(exceptions.NewIllegalParameterException(fmt.Sprintf("watch dog feed lift must more than 0"), nil))
	}
	w := &WatchDog{feeding: life, life: life, callback: callback}
	if err := watchDogList.Offer(w); err != nil {
		err.PrintStackTrace()
		return nil
	}
	return w
}

// Feed feed the dog
func (w *WatchDog) Feed() {
	w.feeding = w.life
}

// Life get the remain life
func (w *WatchDog) Life() int32 {
	return w.feeding
}

// Kill kill this dog
func (w *WatchDog) Kill() {
	w.life = 0
	w.feeding = 0
}
