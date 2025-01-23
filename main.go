package main

import (
	"fmt"
	"github.com/hofcake/go-chip8/chip8"
)

func main() {
	fmt.Println("Hello, World!")

	sys := new(chip8.sys)

	sys.LoadRom("roms/Keypad Test [Hap, 2006].ch8")
	sys.Dump()

}