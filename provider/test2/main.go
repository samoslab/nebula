package main

import (
	"fmt"

	"github.com/eiannone/keyboard"
)

func main() {
	err := keyboard.Open()
	if err != nil {
		panic(err)
	}
	defer keyboard.Close()

	fmt.Println("Press ESC to quit")
	char, _, err := keyboard.GetKey()
	if err != nil {
		panic(err)
	}
	fmt.Printf("You pressed: %q\r\n", char)

}
