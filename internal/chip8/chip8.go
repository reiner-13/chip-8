package chip8

import (
	"fmt"
	"math/rand"
)

type VirtualMachine struct {
	Memory [4096]byte
	IndexRegister uint16
	ProgramCounter uint16
	Stack []uint16
	Registers [16]uint8

	DelayTimer uint8
	SoundTimer uint8

	Display [64][32] byte
	Keyboard [16]byte
}

var fontSet = [80]byte{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

const screenWidth = 64
const screenHeight = 32

func CreateVirtualMachine() *VirtualMachine {
	vm := VirtualMachine{}

	copy(vm.Memory[:], fontSet[:])

	return &vm
}

func (vm *VirtualMachine) Load(data []byte) {
	for i := 0; i < len(data); i++ {
		vm.Memory[512+i] = data[i]
	}
	vm.ProgramCounter = 512
}

func (vm *VirtualMachine) FetchOpcode() uint16 {
	opcode := vm.getOpcode(vm.ProgramCounter)
	vm.incrementPC()
	return opcode
}

func (vm *VirtualMachine) getOpcode(pc uint16) uint16 {
	firstByte := uint16(vm.Memory[pc])
	secondByte := uint16(vm.Memory[pc+1])

	return (firstByte << 8) | secondByte
}

func (vm *VirtualMachine) incrementPC() {
	vm.ProgramCounter += 2
}

func (vm *VirtualMachine) DecodeExecuteOpcode(opcode uint16) {
	instruction := opcode & 0xF000 // 0xX000
	x := (opcode >> 8) & 0x000F // 0x0X00
	y := (opcode >> 4) & 0x000F // 0x00X0
	n := opcode & 0x000F // 0x000X
	nn := opcode & 0x00FF // 0x00XX
	nnn := opcode & 0x0FFF // 0x0XXX
	switch instruction {
	case 0x0000:
		switch nn {
		case 0x00E0:
			fmt.Println("OP: Clear screen")
			for x := 0; x < screenWidth; x++ {
				for y := 0; y < screenHeight; y++ {
					vm.Display[x][y] = 0
				}
			}
		case 0x00EE:
			fmt.Println("OP: Return from subroutine")
			poppedPC := vm.Stack[len(vm.Stack) - 1]
			vm.Stack = vm.Stack[0:len(vm.Stack) - 1]
			vm.ProgramCounter = poppedPC
		default:
			fmt.Println("Unknown opcode.")
		}
	case 0x1000:
		fmt.Printf("OP: Jump to %d\n", nnn)
		vm.ProgramCounter = nnn
	case 0x2000:
		fmt.Printf("OP: Call subroutine at %d\n", nnn)
		vm.Stack = append(vm.Stack, vm.ProgramCounter)
		vm.ProgramCounter = nnn
	case 0x3000:
		fmt.Println("OP: Skip conditionally (3XNN)")
		if vm.Registers[x] == uint8(nn) {
			vm.incrementPC()
		}
	case 0x4000:
		fmt.Println("OP: Skip conditionally (4XNN)")
		if vm.Registers[x] != uint8(nn) {
			vm.incrementPC()
		}
	case 0x5000:
		fmt.Println("OP: Skip conditionally (5XY0)")
		if vm.Registers[x] == vm.Registers[y] {
			vm.incrementPC()
		}
	case 0x6000:
		fmt.Printf("OP: Set register %d to %d\n", x, uint8(nn))
		vm.Registers[x] = uint8(nn)
	case 0x7000:
		fmt.Printf("OP: Add register %d by %d\n", x, uint8(nn))
		vm.Registers[x] += uint8(nn)
	case 0x8000:
		switch n {
		case 0x0000:
			fmt.Printf("OP: Set register %d to register %d\n", x, y)
			vm.Registers[x] = vm.Registers[y]
		case 0x0001:
			fmt.Printf("OP: Binary OR of register %d and register %d\n", x, y)
			vm.Registers[x] = vm.Registers[x] | vm.Registers[y]
		case 0x0002:
			fmt.Printf("OP: Binary AND of register %d and register %d\n", x, y)
			vm.Registers[x] = vm.Registers[x] & vm.Registers[y]
		case 0x0003:
			fmt.Printf("OP: Binary XOR of register %d and register %d\n", x, y)
			vm.Registers[x] = vm.Registers[x] ^ vm.Registers[y]
		case 0x0004:
			fmt.Printf("OP: Add register %d to register %d\n", y, x)
			vx := vm.Registers[x]
			vm.Registers[x] += vm.Registers[y]
			if int(vx) + int(vm.Registers[y]) > 255 {
				vm.Registers[0xF] = 1
			} else {
				vm.Registers[0xF] = 0
			}
		case 0x0005:
			fmt.Printf("OP: Subtract register %d from register %d; store result in register %d\n", y, x, x)
			vx := vm.Registers[x]
			vm.Registers[x] -= vm.Registers[y]
			if vx >= vm.Registers[y] {
				vm.Registers[0xF] = 1
			} else {
				vm.Registers[0xF] = 0
			}
		case 0x0006:
			fmt.Printf("OP: Shift right register %d\n", x)
			carryBit := vm.Registers[x] & 0x01
			vm.Registers[x] = vm.Registers[x] >> 1
			vm.Registers[0xF] = carryBit
		case 0x0007:
			fmt.Printf("OP: Subtract register %d from register %d; store result in register %d\n", x, y, x)
			vx := vm.Registers[x]
			vm.Registers[x] = vm.Registers[y] - vm.Registers[x]
			if vm.Registers[y] >= vx {
				vm.Registers[0xF] = 1
			} else {
				vm.Registers[0xF] = 0
			}
		case 0x000E:
			fmt.Printf("OP: Shift left register %d\n", x)
			carryBit := (vm.Registers[x] >> 7) & 0x01
			vm.Registers[x] = vm.Registers[x] << 1
			vm.Registers[0xF] = carryBit
		default:
			fmt.Println("Unknown opcode.")
		}
	case 0x9000:
		fmt.Println("OP: Skip conditionally (9XY0)")
		if vm.Registers[x] != vm.Registers[y] {
			vm.incrementPC()
		}
	case 0xA000:
		fmt.Printf("OP: Set index register to %d\n", nnn)
		vm.IndexRegister = nnn
	case 0xB000:
		fmt.Printf("OP: Jump to %d + %d\n", nnn, vm.Registers[0])
		vm.ProgramCounter = nnn + uint16(vm.Registers[0])
	case 0xC000:
		fmt.Println("OP: Random")
		vm.Registers[x] = uint8(rand.Intn(256)) & uint8(nn)
	case 0xD000:
		fmt.Println("OP: Draw sprite")
		xCoord := vm.Registers[x]
		yCoord := vm.Registers[y]
		vm.Registers[0xF] = 0

		for byteIndex := uint16(0); byteIndex < n; byteIndex++ {
			spriteRow := vm.Memory[vm.IndexRegister + byteIndex]
			for bitIndex := 0; bitIndex < 8; bitIndex++ {
				displayX := (uint16(xCoord) + uint16(bitIndex)) % screenWidth
				displayY := (uint16(yCoord) + byteIndex) % screenHeight

				if (displayX >= screenWidth || displayY >= screenHeight) {
					continue
				}

				spriteBit := (spriteRow >> (8 - bitIndex - 1)) & 0x1
				displayBit := &vm.Display[displayX][displayY]

				if spriteBit == 1 && *displayBit == 1 {
					vm.Registers[0xF] = 1
				}
				*displayBit = *displayBit ^ spriteBit
			}
		}
	case 0xE000:
		switch nn {
		case 0x009E:
			fmt.Println("OP: Skip if key is pressed")
			if vm.Keyboard[vm.Registers[x]] == 1 {
				vm.incrementPC()
			}
		case 0x00A1:
			fmt.Println("OP: Skip if key is pressed")
			if vm.Keyboard[vm.Registers[x]] != 1 {
				vm.incrementPC()
			}
		default:
			fmt.Println("Unknown opcode.")
		}
	case 0xF000:
		switch nn {
		case 0x0007:
			fmt.Printf("OP: Set register %d to value of delay timer\n", x)
			vm.Registers[x] = vm.DelayTimer
		case 0x0015:
			fmt.Printf("OP: Set delay timer to value of register %d\n", x)
			vm.DelayTimer = vm.Registers[x]
		case 0x0018:
			fmt.Printf("OP: Set sound timer to value of register %d\n", x)
			vm.SoundTimer = vm.Registers[x]
		case 0x001E:
			fmt.Printf("OP: Add register %d to index register\n", x)
			vm.IndexRegister += uint16(vm.Registers[x])
		case 0x000A:
			fmt.Println("OP: Get key")
			keyPress := false
			for i := range len(vm.Keyboard) {
				if vm.Keyboard[i] == 1 {
					vm.Registers[x] = uint8(i)
					keyPress = true
				}
			}
			if !keyPress {
				vm.ProgramCounter -= 2
			}
		case 0x0029:
			fmt.Printf("OP: Index register set to address of font character in register %d\n", x)
			vm.IndexRegister = uint16(vm.Memory[0 + vm.Registers[x]])
		case 0x0033:
			fmt.Println("OP: Binary-coded decimal conversion")
			vm.Memory[vm.IndexRegister] = vm.Registers[x] / 100
			vm.Memory[vm.IndexRegister + 1] = (vm.Registers[x] / 10) % 10
			vm.Memory[vm.IndexRegister + 2] = (vm.Registers[x] % 100) % 10
		case 0x0055:
			fmt.Printf("OP: Store up to register %d in memory\n", x)
			for i := uint16(0); i <= x; i++ {
				vm.Memory[vm.IndexRegister + i] = vm.Registers[i]
			}
		case 0x0065:
			fmt.Println("OP: Load values into registers")
			for i := uint16(0); i <= x; i++ {
				vm.Registers[i] = vm.Memory[vm.IndexRegister + i]
			}
		default:
			fmt.Println("Unknown opcode.")
		}
	default:
		fmt.Println("Unknown opcode.")
	}
}

func (vm *VirtualMachine) Tick() {
	if (vm.DelayTimer > 0) {
		vm.DelayTimer -= 1
	}
	if (vm.SoundTimer > 0) {
		vm.SoundTimer -= 1
	}
	if (vm.SoundTimer == 0) {
		fmt.Println("BEEP")
	}
}