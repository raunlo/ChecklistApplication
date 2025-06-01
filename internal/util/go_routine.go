package util

import (
	"fmt"
	"sync"
)

type Func[ReturnType any] func() ReturnType

func GoroutineWrapperFunction[ReturnType any](fn Func[ReturnType], wg *sync.WaitGroup, ch chan<- ReturnType) {
	res := fn()
	fmt.Printf("%+v", res)
	wg.Done()
	ch <- res
}
