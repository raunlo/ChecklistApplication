package util

import (
	"sync"
)

type Func[ReturnType any] func() ReturnType

func GoroutineWrapperFunction[ReturnType any](fn Func[ReturnType], wg *sync.WaitGroup, ch chan<- ReturnType) {
	res := fn()
	wg.Done()
	ch <- res
}
