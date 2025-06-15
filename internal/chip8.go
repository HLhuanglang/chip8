package chip8

type RAM struct {
	/*
		内存布局说明：
		0x000 ~ 0x1ff：解释器               0    ~ 511    512字节
		0x200 ~ 0xe9f：程序可自由使用       512  ~ 3743   3232字节
		0xea0 ~ 0xeff：保留给栈及内部应用    3744 ~ 3839   96字节
		0xf00 ~ 0xfff：保留给屏幕使用        3840 ~ 4095   256字节 = 64x32分辨率(bit)
	*/
	Memory [4096]byte
}

type CPU struct {
	PC    uint16   //program counter
	SP    byte     //stack pointer
	Stack [16]byte //stack
	VRegs [16]byte //8bit register：V0~VF
	IReg  uint16   //16bit register: I, used to store memory addresses
	STReg byte     //sound timers
	DTReg byte     //delay
}

type Chip8Machine struct {
	Cpu *CPU
	Mem *RAM
}
