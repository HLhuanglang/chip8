package main

import chip8 "github.com/HLhuanglang/chip8/internal"

func main() {
	vm := chip8.NewChip8Machine()
	err := vm.Init()
	if err != nil {
		panic(err)
	}
	vm.Run()
}
