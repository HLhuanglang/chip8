package chip8

/*
指令2字节长度：大端(数值高位存储在内存的低地址，低位存储在内存的高地址)
0000 0000 0000 0000
	   N	N	N	：12-bit value, the lowest 12 bits of the instruction 表示12位内存地址(2^12 = 4096字节，12位可以寻址到全部地址)
	   		N	N	：An 8-bit value, the lowest 8 bits of the instruction
	   			N	：A 4-bit value, the lowest 4 bits of the instruction
	   X 			：A 4-bit value, the lower 4 bits of the high byte of the instruction
	   		Y		: A 4-bit value, the upper 4 bits of the low byte of the instruction
*/

type Instruction struct {
	Opcode uint16
	NNN    uint16
	NN     uint8
	N      uint8
	X      uint8
	Y      uint8
}

func NewInstruction(opcode uint16) Instruction {
	return Instruction{
		Opcode: opcode,
		NNN:    opcode & 0x0fff,
		NN:     uint8(opcode) & 0x0ff,
		N:      uint8(opcode) & 0x0f,
		X:      uint8(opcode>>8) & 0x0f,
		Y:      uint8(opcode>>4) & 0x0f,
	}
}
