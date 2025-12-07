package main

import "fmt"

type duration int

func newDuration(d int) duration {
	return duration(d)
}

func (d *duration) Print() {
	fmt.Println("duration is ", *d)
}

type greeter struct {
	name string
}

func (g *greeter) Print() {
	fmt.Println("greeting ", g.name)
}

func main() {
	// Because 10 is a constant, it doesn't have a memory address, so
	// trying to invoke a pointer receiver of `duration` will result in
	// a compile error of `cannot call pointer method Print on duration`.
	// duration(10).Print()
	// (&newDuration(10)).Print()

	// You can invoke pointer method once you assigned the duration value
	// to a variable though.
	d := duration(10)
	d.Print()

	// You can invoke a pointer method on the address of a struct literal though.
	(&greeter{name: "John"}).Print()
}
