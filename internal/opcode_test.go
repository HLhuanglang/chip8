package chip8

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpcode(t *testing.T) {
	isr := NewInstruction(0x22fc)
	assert.Equal(t, uint16(0x22fc), isr.Opcode)
	assert.Equal(t, uint16(0x2fc), isr.NNN)
	assert.Equal(t, uint8(0xfc), isr.NN)
	assert.Equal(t, uint8(0xc), isr.N)
	assert.Equal(t, uint8(0x2), isr.X)
	assert.Equal(t, uint8(0xf), isr.Y)
}
