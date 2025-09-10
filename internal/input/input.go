package input

import (
	"chip8emu/internal/chip8"

	"github.com/veandco/go-sdl2/sdl"
)

var keyMap = map[sdl.Scancode]byte{
	sdl.SCANCODE_1: 0x1,
	sdl.SCANCODE_2: 0x2,
	sdl.SCANCODE_3: 0x3,
	sdl.SCANCODE_4: 0xC,
	sdl.SCANCODE_Q: 0x4,
	sdl.SCANCODE_W: 0x5,
	sdl.SCANCODE_E: 0x6,
	sdl.SCANCODE_R: 0xD,
	sdl.SCANCODE_A: 0x7,
	sdl.SCANCODE_S: 0x8,
	sdl.SCANCODE_D: 0x9,
	sdl.SCANCODE_F: 0xE,
	sdl.SCANCODE_Z: 0xA,
	sdl.SCANCODE_X: 0x0,
	sdl.SCANCODE_C: 0xB,
	sdl.SCANCODE_V: 0xF,
}

func InputListen(vm *chip8.VirtualMachine) {
	for {
		event := sdl.PollEvent()
		if event == nil {
			continue
		}

		
	}
}