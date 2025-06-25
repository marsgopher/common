package concurrency

import (
	"fmt"
)

func ExampleSemaWaitGroup() {
	g := NewSemaWaitGroup(2)
	var items = []string{
		"aaa",
		"bbb",
		"ccc",
		"ddd",
	}
	m := make([]string, len(items))
	for i, item := range items {
		i, item := i, item
		g.Do(func() {
			m[i] = item
		})
	}
	g.Wait()
	fmt.Println(m)

	// Output:
	// [aaa bbb ccc ddd]
}
