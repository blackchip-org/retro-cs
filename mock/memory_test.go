package mock

import (
	"fmt"
)

func ExampleMockRead() {
	data := []int{1, 2, 3}
	read := MockRead(data)
	fmt.Println(read())
	fmt.Println(read())
	fmt.Println(read())
	// Output:
	// 1
	// 2
	// 3
}
