package main



type sys struct {
	vx [16]uint8
	i, pc uint16
	sp, dt, st uint8

	ram [4096]uint8
	gfx [64 * 32]uint8
	keys [16]uint8

}

func (s *sys) Dump(){
	fmt.println("Vx: ", s.vx)
	fmt.println("I: ", s.i)
	fmt.println("PC: ", s.pc)
	fmt.println("SP: ", s.sp)
	fmt.println("Memory")
	fmt.println(s.ram)
}

func (s *sys) LoadRom(rom []uint8){

	// decide on rom format, there are some nice roms that have headers?
}

func (s *sys) Step(){
	// fetch opcode
	// decode opcode
	// execute opcode
}

