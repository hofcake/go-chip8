package chip8

import (
	"fmt"
	"os"
	"log"
) 


type sys struct {
	vx [16]uint8
	i, pc uint16
	sp, dt, st uint8

	ram [4096]uint8
	gfx [64 * 32]uint8
	keys [16]uint8

}

func (s *sys) Dump(){
	fmt.Println("Vx: ", s.vx)
	fmt.Println("I: ", s.i)
	fmt.Println("PC: ", s.pc)
	fmt.Println("SP: ", s.sp)
	fmt.Println("Memory")
	fmt.Println(s.ram)
}

func (s *sys) LoadRom(rom string){


	data, err := os.ReadFile(rom)
	if err != nil {
		log.Fatal(err)
	}

	if len(data) > 4096 - 512 {
		log.Fatal("Rom too big")
	}

	n := copy(s.ram[512:], data)
	log.Println("loaded ", n, "bytes")
	
	
}

func (s *sys) Step(){
	// fetch opcode
	// decode opcode
	// execute opcode
}

func (s *sys) fetch(){
	// fetch opcode
}

func (s *sys) decode(){
	// decode opcode
}

func (s *sys) execute(){
	// execute opcode
}