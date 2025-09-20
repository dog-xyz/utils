package sds

import (
	"fmt"
	"sync"
	"testing"
)

func TestSync(t *testing.T) {
	var once sync.Once
	wg := sync.WaitGroup{}
	wg.Add(5)

	for i := 0; i < 5; i++ {
		go func(wg *sync.WaitGroup, i int) {
			defer wg.Done()
			fmt.Printf("run %d\n", i)
			once.Do(func() {
				fmt.Printf("hello\n")
			})

		}(&wg, i)
	}

	wg.Wait()
}
