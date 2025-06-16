package chip8

import (
	"fmt"
	"time"

	"github.com/ncruces/zenity"
	"github.com/veandco/go-sdl2/sdl"
)

type Chip8Machine struct {
	/*
		内存布局说明：
		0x000 ~ 0x1ff：解释器               0    ~ 511    512字节
		0x200 ~ 0xe9f：程序可自由使用       512  ~ 3743   3232字节
		0xea0 ~ 0xeff：保留给栈及内部应用    3744 ~ 3839   96字节
		0xf00 ~ 0xfff：保留给屏幕使用        3840 ~ 4095   256字节 = 64x32分辨率(bit)
	*/
	Memory [4096]byte
	PC     uint16   //program counter
	Stack  [16]byte //stack：由于存在跳转指令,因此需要
	SP     byte     //stack pointer
	VxRegs [16]byte //8bit register：V0~VF, VF是特殊的寄存器
	IReg   uint16   //16bit register: I, used to store memory addresses
	STReg  byte     //非0则激活sound timer，以60hz的频率递减
	DTReg  byte     //非0则激活delay timer，以60hz的频率递减

	RomPath string //存储当前运行rom的路径
	Window  *sdl.Window
	Render  *sdl.Renderer
}

func NewChip8Machine() *Chip8Machine {
	return &Chip8Machine{}
}

func (c *Chip8Machine) Init() error {
	err := sdl.Init(sdl.INIT_AUDIO | sdl.INIT_VIDEO | sdl.INIT_EVENTS)
	if err != nil {
		return err
	}
	c.Window, err = sdl.CreateWindow("chip8", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 1024, 1024, sdl.WINDOW_SHOWN)
	if err != nil {
		return err
	}
	c.Render, err = sdl.CreateRenderer(c.Window, -1, 0)
	if err != nil {
		return err
	}
	return nil
}

func (c *Chip8Machine) Run() {
	cpu_tick := time.NewTicker(time.Millisecond)   //1khz
	audio_tick := time.NewTicker(time.Second / 60) //60hz
	gpu_tick := time.NewTicker(time.Second / 60)   //60hz
	for c.handleUserinput() {
		select {
		case <-cpu_tick.C:
			//一个指令周期,需要读取+解码+运算
			c.emulateInstruction()
		case <-audio_tick.C:
			//声音，只有bee的一下响，没有音乐
			c.updateVideo()
		case <-gpu_tick.C:
			//渲染图像数据到屏幕
			c.updateScreen()
		default:
			//do nothing
		}
	}
	c.Window.Destroy()
	sdl.Quit()
}

func (c *Chip8Machine) handleUserinput() bool {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.QuitEvent:
			return false
		case *sdl.MouseMotionEvent:
			fmt.Printf("mouse mov:(%d,%d)\n", t.X, t.Y)
		case *sdl.KeyboardEvent:
			switch t.Keysym.Scancode {
			case sdl.SCANCODE_F2:
				if t.Type == sdl.KEYDOWN {
					path, err := zenity.SelectFile()
					if err != nil {
						fmt.Printf("err=%+v\n", err)
						if err == zenity.ErrCanceled {
							fmt.Printf("cancel selectfile\n")
						}
					} else {
						c.RomPath = path
					}
				}
			}
		default:
			//do nothing
		}
	}
	return true
}

func (c *Chip8Machine) emulateInstruction() {
	//TODO
}

func (c *Chip8Machine) updateScreen() {
	//TODO
}

func (c *Chip8Machine) updateVideo() {
	//TODO
}
