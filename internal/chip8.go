package chip8

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ncruces/zenity"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	TotalRAMSize      = 4096
	MaxRomSize        = 3584
	RegisterSize      = 16
	RomEntryPoint     = 0x200
	ETIRomEntoryPoint = 0x600
	WindowName        = "chip8"
	WindowHeight      = 470
	WindowWidth       = 800
)

type MachineStatus int

const (
	SYS_PAUSE MachineStatus = iota
	SYS_RUNNING
	SYS_QUIT
)

type Chip8Machine struct {
	/*
		内存布局说明：
		0x000 ~ 0x1ff：解释器               0    ~ 511    512字节
		0x200 ~ 0xe9f：程序可自由使用       512  ~ 3743   3232字节
		0xea0 ~ 0xeff：保留给栈及内部应用    3744 ~ 3839   96字节
		0xf00 ~ 0xfff：保留给屏幕使用        3840 ~ 4095   256字节 = 64x32分辨率(bit)
	*/
	Memory [TotalRAMSize]byte
	PC     uint16             //program counter
	Stack  [16]byte           //stack：由于存在跳转指令,因此需要
	SP     byte               //stack pointer
	VxRegs [RegisterSize]byte //8bit register：V0~VF, VF是特殊的寄存器
	IReg   uint16             //16bit register: I, used to store memory addresses
	STReg  byte               //非0则激活sound timer，以60hz的频率递减
	DTReg  byte               //非0则激活delay timer，以60hz的频率递减

	State         MachineStatus
	RomPath       string //存储当前运行rom的路径
	RomEntryPoint uint16
	Window        *sdl.Window
	Render        *sdl.Renderer
	Screnn        *sdl.Texture
	Font          *sdl.Texture
}

type MachineOption struct {
	UseETI bool
}

type OptionFunc func(*MachineOption)

func WithETI() OptionFunc {
	return func(m *MachineOption) {
		m.UseETI = true
	}
}

func NewChip8Machine(opts ...OptionFunc) *Chip8Machine {
	machineOpt := &MachineOption{}
	for _, opt := range opts {
		opt(machineOpt)
	}
	entryPoint := RomEntryPoint
	if machineOpt.UseETI {
		entryPoint = ETIRomEntoryPoint
	}
	return &Chip8Machine{
		RomEntryPoint: uint16(entryPoint),
		State:         SYS_PAUSE,
	}
}

func (c *Chip8Machine) Init() error {
	err := sdl.Init(sdl.INIT_AUDIO | sdl.INIT_VIDEO | sdl.INIT_EVENTS)
	if err != nil {
		return err
	}
	c.Window, err = sdl.CreateWindow(WindowName, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, WindowWidth, WindowHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		return err
	}
	c.Render, err = sdl.CreateRenderer(c.Window, -1, 0)
	if err != nil {
		return err
	}
	fmt := sdl.PIXELFORMAT_RGB888
	access := sdl.TEXTUREACCESS_TARGET
	c.Screnn, err = c.Render.CreateTexture(uint32(fmt), access, 8, 8)
	if err != nil {
		return err
	}
	c.loadFont()
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
		case *sdl.KeyboardEvent:
			switch t.Keysym.Scancode {
			case sdl.SCANCODE_F1:
				if t.Type == sdl.KEYDOWN {
					c.State = SYS_PAUSE
				}
			case sdl.SCANCODE_F2:
				if t.Type == sdl.KEYDOWN {
					c.loadRom()
				}
			}
		default:
			//do nothing
		}
	}
	return true
}

func (c *Chip8Machine) loadRom() {
	path, err := zenity.SelectFile()
	if err != nil {
		fmt.Printf("err=%+v\n", err)
		if err == zenity.ErrCanceled {
			fmt.Printf("cancel selectfile\n")
		}
	} else {
		c.RomPath = path
		file, err := os.Open(path)
		if err != nil {
			fmt.Printf("failed to open rom: %v\n", err)
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil {
			fmt.Printf("get fileinfo faild:%v\n", err)
		}
		if info.Size() > MaxRomSize {
			fmt.Printf("Rom size to big")
		}
		n, err := io.ReadFull(file, c.Memory[0x200:])
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			fmt.Printf("failed to read rom: %v\n", err)
		}
		fmt.Printf("loaded rom: %s (%d bytes)\n", path, n)
		c.PC = c.RomEntryPoint
		c.State = SYS_RUNNING
	}
}

func (c *Chip8Machine) emulateInstruction() {
	if c.State != SYS_RUNNING {
		return

	}
	inst := NewInstruction(c.fetch())
	if c.PC > 0xfff {
		c.State = SYS_QUIT
	}
	fmt.Printf("Address:0x%X Opcode:0x%X NNN:0x%x NN:0x%x N:0x%x X:%d Y:%d\n", c.PC-2, inst.Opcode, inst.NNN, inst.NN, inst.N, inst.X, inst.Y)
	opType := inst.Opcode >> 12
	switch opType { //指令集分类见：https://github.com/mattmikolay/chip-8/wiki/CHIP%E2%80%908-Instruction-Set
	case 0x0:
		{
		}
	case 0x1:
		{
		}
	case 0x2:
		{
		}
	case 0x3:
		{
		}
	case 0x4:
		{
		}
	case 0x5:
		{
		}
	case 0x6:
		{
		}
	case 0x7:
		{
		}
	case 0x8:
		{
		}
	case 0x9:
		{
		}
	case 0xA:
		{
		}
	case 0xB:
		{
		}
	case 0xC:
		{
		}
	case 0xD:
		{
		}
	case 0xE:
		{
		}
	case 0xF:
		{
		}
	}
}

func (c *Chip8Machine) fetch() uint16 {
	opt1 := uint16(c.Memory[c.PC]) << 8 //cmt：Memory类型是byte,导致<<8清空,需要转成uint16, 浪费调试时间1小时...
	opt2 := uint16(c.Memory[c.PC+1])
	opcode := opt1 | opt2
	c.PC += 2
	return opcode
}

func (c *Chip8Machine) updateScreen() {
	//绘制背景
	c.Render.SetDrawColor(32, 42, 53, 255)
	c.Render.Clear()

	//绘制内部边框,划分四大块区域,边框线条占用1pt
	c.splitRegion(8, 8, 512+2, 256+2)     //游戏屏幕
	c.splitRegion(526, 8, 266+2, 256+2)   //PC-地址-指令集
	c.splitRegion(8, 270, 512+2, 192+2)   //日志信息打印
	c.splitRegion(526, 270, 266+2, 192+2) //寄存器/栈

	//绘制不同区域内容
	c.drawScreen()
	c.drawInst()
	c.drawLog()
	c.drawRegs()

	//渲染
	c.Render.Present()
}

func (c *Chip8Machine) splitRegion(x, y, w, h int32) {
	c.Render.SetDrawColor(0, 0, 0, 255)
	c.Render.DrawLine(x, y, x+w, y) //上
	c.Render.DrawLine(x, y, x, y+h) //左
	c.Render.SetDrawColor(95, 112, 120, 255)
	c.Render.DrawLine(x+w, y, x+w, y+h) //右
	c.Render.DrawLine(x, y+h, x+w, y+h) //下
}

func (c *Chip8Machine) drawScreen() {
	src := sdl.Rect{
		W: int32(8),
		H: int32(8),
	}

	// stretch the render target to fit
	c.Render.Copy(c.Screnn, &src, &sdl.Rect{X: 8, Y: 8, W: 512, H: 256})
}

func (c *Chip8Machine) drawInst() {
	//TODO
}

func (c *Chip8Machine) drawLog() {
	c.drawText("ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890", 8+2, 270+2)
}

func (c *Chip8Machine) drawRegs() {
	//TODO
}

func (c *Chip8Machine) updateVideo() {
	//TODO
}

func (c *Chip8Machine) loadFont() {
	var surface *sdl.Surface
	var err error

	if surface, err = sdl.LoadBMP("font.bmp"); err != nil {
		panic(err)
	}

	// get the magenta color
	mask := sdl.MapRGB(surface.Format, 255, 0, 255)

	// set the mask color key
	surface.SetColorKey(true, mask)

	// create the texture
	if c.Font, err = c.Render.CreateTextureFromSurface(surface); err != nil {
		panic(err)
	}
}

func (c *Chip8Machine) drawText(s string, x, y int) {
	src := sdl.Rect{W: 5, H: 7}
	dst := sdl.Rect{
		X: int32(x),
		Y: int32(y),
		W: 5,
		H: 7,
	}

	// loop over all the characters in the string
	for _, ch := range s {
		if ch > 32 && ch < 127 {
			src.X = (ch - 33) * 6

			// draw the character to the renderer
			c.Render.Copy(c.Font, &src, &dst)
		}

		// advance
		dst.X += 7
	}
}
