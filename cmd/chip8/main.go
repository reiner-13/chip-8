package main

import (
	"chip8emu/internal/chip8"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	screenWidth = 64
	screenHeight = 32
	scale = 10
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

func main() {
	fmt.Printf("Starting Chip-8...\n")

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		log.Fatalf("Failed to initialize SDL: %v", err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow(
		"Chip-8 Emulator",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		screenWidth * scale,
		screenHeight * scale,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		log.Fatalf("Failed to create window: %v", err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		log.Fatalf("Failed to create renderer: %v", err)
	}
	defer renderer.Destroy()

	vm := chip8.CreateVirtualMachine()

	loadROM(vm, "./roms/pumpkindressup.ch8")
	
	run(vm, renderer)
}

func loadROM(vm *chip8.VirtualMachine, filePath string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading ROM: %v", err)
	}
	vm.Load(data)
}

func run(vm *chip8.VirtualMachine, renderer *sdl.Renderer) {
	fmt.Println("RUN")
	go startMachine(vm)
	go startTimers(vm)
	startDisplay(vm, renderer)
}

func startMachine(vm *chip8.VirtualMachine) {
	ticker := time.NewTicker(time.Second / 700) // 700 Hz
	defer ticker.Stop()
	
	for range ticker.C {
		opcode := vm.FetchOpcode()
		vm.DecodeExecuteOpcode(opcode)
	}
}

func startTimers(vm *chip8.VirtualMachine) {
	ticker := time.NewTicker(time.Second / 60) // 60 Hz
	defer ticker.Stop()
	
	for range ticker.C {
		vm.Tick()
	}
}

func startDisplay(vm *chip8.VirtualMachine, renderer *sdl.Renderer) {
	ticker := time.NewTicker(time.Second / 60) // 60 Hz
	defer ticker.Stop()
	
	running := true
	for range ticker.C {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				switch e.Type {
				case sdl.KEYDOWN:
					if key, ok := keyMap[e.Keysym.Scancode]; ok {
						vm.Keyboard[key] = 0x01
					}
				case sdl.KEYUP:
					if key, ok := keyMap[e.Keysym.Scancode]; ok {
						vm.Keyboard[key] = 0x00
					}
				}
			}
		}
		if !running {
			break
		}

		renderer.SetDrawColor(100, 140, 220, 255)
		renderer.Clear()

		renderer.SetDrawColor(20, 20, 60, 255)
		for x := 0; x < screenWidth; x++ {
			for y := 0; y < screenHeight; y++ {
				if vm.Display[x][y] == 1 {
					rect := sdl.Rect{
						X: int32(x * scale),
						Y: int32(y * scale),
						W: scale,
						H: scale,
					}
					renderer.FillRect(&rect)
				}
			}
		}

		renderer.Present()


	}
}