package main

import (
	"fmt"
)

const hello_prefix = "Hello, "

func Hello(name string) string {

	if name == "" {
		name = "World"
	}

	return hello_prefix + name
}

func GenerateScore(score bool) int {
	if score {
		return 100
	} else {
		return 50
	}
}

func main() {
	var name = "Zack"
	fmt.Printf("Hello, %v \n", name)

	test := Hello("Zack")
	fmt.Printf("%v\n", test)
}
