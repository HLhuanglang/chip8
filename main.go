package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	// 初始化：chip8只需要渲染+音频,无需其他模块
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	// 创建窗口
	var window *sdl.Window
	window, err := sdl.CreateWindow("chip8", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, 1024, 1024, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	//创建render
	var render *sdl.Renderer
	render, err = sdl.CreateRenderer(window, -1, 0)
	if err != nil {
		panic(err)
	}
	defer render.Destroy()

	//创建surface
	var surf *sdl.Surface
	surf, err = sdl.LoadBMP("./assets/chip-8_16.bmp")
	if err != nil {
		panic(err)
	}
	defer surf.Free()

	//创建texture
	texture, err := render.CreateTextureFromSurface(surf)
	if err != nil {
		panic(err)
	}
	defer texture.Destroy()

	//渲染
	src := sdl.Rect{0, 0, 48, 48}
	dst := sdl.Rect{50, 50, 48, 48}
	render.Clear()
	render.Copy(texture, &src, &dst)
	render.Present()

	var event sdl.Event
	running := true
	for running {
		for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			sdlT := event.GetType()
			switch sdlT {
			case sdl.QUIT:
				fmt.Print("quit~")
				running = false
			}
		}
		sdl.Delay(10)
	}
}
