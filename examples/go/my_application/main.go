package main

import (
	"fmt"
	"time"
)

func main() {
	hello := 0
	for {
		hello++
		fmt.Printf("I've said hello %d time(s)\n", hello)
		time.Sleep(2 * time.Second)

		if hello == 100 {
			fmt.Println("I'm tired, goodbye")
			break
		}
	}
}
