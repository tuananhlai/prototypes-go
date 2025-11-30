package main

import "fmt"

type duration int

func (d *duration) Print() {
	fmt.Println("duration is ", *d)
}

func main() {
	d := duration(10)
	d.Print()

	// Because 10 is a constant, it doesn't have a memory address, so
	// trying to invoke a pointer receiver of `duration` will result in
	// a compile error of `cannot call pointer method Print on duration`.
	// duration(10).Print()
}
