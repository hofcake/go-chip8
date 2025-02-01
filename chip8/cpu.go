package chip8

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	// "math/bits"
)

type Sys struct {

	// CHIP8 Spec
	vn         [16]uint8
	i, pc      uint16
	sp, dt, st uint8

	ram   [4096]uint8
	stack [16]uint16
	gfx   [64 * 32]uint8
	keys  [16]uint8

	// Additional
	sop, eop uint16
}

// Eventually add the start of program adress as a parameter to increase compatibility
func InitSys() *Sys {
	s := new(Sys)
	s.pc = 512
	s.sop = 512
	return s
}

func (s *Sys) Dump() {
	fmt.Println("Vn: ", s.vn)
	fmt.Println("I: ", s.i)
	fmt.Println("PC: ", s.pc)
	fmt.Println("SP: ", s.sp)
	fmt.Println("Memory")
	fmt.Println(s.ram)

}

// Returns readable assembly from start of program to end of program
func (s *Sys) Disasm() []string {

	var disasm = make([]string, int(s.eop-s.sop)/2)

	fmt.Printf("reading from %d to %d, dissasm len: %d\n", s.sop, s.eop, len(disasm))

	j := 0
	for i := s.sop; i < s.eop; i += 2 {
		instr := uint16(s.ram[i])<<8 | uint16(s.ram[i+1])
		disasm[j] = decode(instr)
		j++
	}

	return disasm

}

func (s *Sys) PrintNext() {

	var instr uint16 = uint16(s.ram[s.pc])<<8 | uint16(s.ram[s.pc+1])

	fmt.Printf("%#04x\n", instr)
	decode(instr)

	s.pc += 2
}

func (s *Sys) LoadRom(rom string) {

	data, err := os.ReadFile(rom)
	if err != nil {
		log.Fatal(err)
	}

	if len(data) > len(s.ram)-int(s.sop) {
		log.Fatal("Rom too big")
	}

	n := copy(s.ram[s.sop:], data)
	s.eop = uint16(n + int(s.sop))
	log.Println("loaded ", n, "bytes")

}

func (s *Sys) Step() {
	// fetch opcode
	// decode opcode
	// execute opcode
}

func (s *Sys) execute(instr uint16) {
	d1 := instr & 0xF000 >> 12
	// d2 := instr & 0x0F00 >> 8
	d3 := instr & 0x00F0 >> 4
	d4 := instr & 0x000F

	nnn := instr & 0x0FFF
	kk := uint8(instr & 0x00FF)
	// n := instr & 0x000F

	x := instr & 0x0F00 >> 8 // redundant
	y := instr & 0x00F0 >> 8 // redundant

	var temp uint16

	switch {

	case instr == 0x00E0: // clears screan
		s.gfx = [len(s.gfx)]uint8{}
	case instr == 0x00EE: // returns from subroutine
		s.pc = s.stack[s.sp]
		s.sp--
	case d1 == 0x1: // jump to addr nnn
		s.pc = nnn
	case d1 == 0x2: // call addr nnn
		s.sp++
		s.stack[s.sp] = nnn
	case d1 == 0x3: // Skip next if Vx = kk
		if s.vn[x] == kk {
			s.pc += 2
		}
	case d1 == 0x4: // Skip next if vx != kk
		if s.vn[x] != kk {
			s.pc += 2
		}
	case d1 == 0x5: // Skip next if vx == vy
		if s.vn[x] == s.vn[y] {
			s.pc += 2
		}
	case d1 == 0x6: // set Vx = kk
		s.vn[x] = kk
	case d1 == 0x7: // set Vx = vx + kk
		s.vn[x] += kk
	case d1 == 0x8 && d4 == 0x0: // vx = vy
		s.vn[x] = s.vn[y]
	case d1 == 0x8 && d4 == 0x1: // vx = vx | vy
		s.vn[x] |= s.vn[y]
	case d1 == 0x8 && d4 == 0x2: // vx = vx & vy
		s.vn[x] &= s.vn[y]
	case d1 == 0x8 && d4 == 0x3: // vx = vx xor vy
		s.vn[x] ^= s.vn[y]
	case d1 == 0x8 && d4 == 0x4: // vx = vx + vy, vf = carry
		temp = uint16(s.vn[x]) + uint16(s.vn[y])
		if temp > 255 {
			s.vn[15] = 1
		}
		s.vn[x] = uint8(temp & 0x00FF)
	case d1 == 0x8 && d4 == 0x5: // vx = vx - vy, vf = not borrow
		if s.vn[x] > s.vn[y] {
			s.vn[15] = 1
		} else {
			s.vn[15] = 0

		}
		s.vn[x] -= s.vn[y]
	case d1 == 0x8 && d4 == 0x6: // vx = vx shr 1
		if s.vn[x]%2 == 1 {
			s.vn[15] = 1
			// do we still divide by 2? YESSSSSS
		} else {
			s.vn[15] = 0
		}
		s.vn[x] /= 2
	case d1 == 0x8 && d4 == 0x7: // vx = vy - vx, vf = not borrow
		if s.vn[y] > s.vn[x] {
			s.vn[15] = 1
			// do we still subtract? YES
		} else {
			s.vn[15] = 0
			s.vn[x] = s.vn[y] - s.vn[x]
		}
	case d1 == 0x8 && d4 == 0xE: // vx = vx shl 1
		if s.vn[x]&(1<<7) == 1 {
			s.vn[15] = 1
		} else {
			s.vn[15] = 0
			s.vn[x] *= 2
		}
	case d1 == 0x9:
		if s.vn[x] != s.vn[y] {
			s.pc += 2
		}
	case d1 == 0xa:
		s.i = nnn
	case d1 == 0xb:
		s.pc = nnn + uint16(s.vn[0])
	case d1 == 0xc:
		s.vn[x] = uint8(rand.Intn(255)) & kk
	case d1 == 0xd:
		// implement display functionality here first
	case d1 == 0xe && d4 == 0xe:
	case d1 == 0xe && d4 == 0x1:
	case d1 == 0xf && d4 == 0x7:
	case d1 == 0xf && d4 == 0xa:
	case d1 == 0xf && d3 == 0x1 && d4 == 0x5:
	case d1 == 0xf && d4 == 0x8:
	case d1 == 0xf && d4 == 0xe:
	case d1 == 0xf && d4 == 0x9:
	case d1 == 0xf && d4 == 0x3:
	case d1 == 0xf && d3 == 0x5 && d4 == 0x5:
	case d1 == 0xf && d3 == 0x6 && d4 == 0x5:
	default:
	}
}

func decode(instr uint16) string {
	assm := ""

	d1 := instr & 0xF000 >> 12
	// d2 := instr & 0x0F00 >> 8
	d3 := instr & 0x00F0 >> 4
	d4 := instr & 0x000F

	nnn := instr & 0x0FFF
	kk := instr & 0x00FF
	// n = instr & 0x000F

	x := instr & 0x0F00 >> 8 // redundant
	y := instr & 0x00F0 >> 8 // redundant

	// var temp uint16

	switch {

	case instr == 0x00E0: // clears screan
		assm = "CLS"
	case instr == 0x00EE: // returns from subroutine
		assm = "RET"
	case d1 == 0x1: // jump to addr nnn
		assm = fmt.Sprintf("JP %#03x", nnn)
	case d1 == 0x2: // call addr nnn
		assm = fmt.Sprintf("CALL %#03x", nnn)
	case d1 == 0x3: // Skip next if Vx = kk
		assm = fmt.Sprintf("SE V%d, %d", x, kk)
	case d1 == 0x4: // Skip next if vx != kk
		assm = fmt.Sprintf("SNE V%d, %d", x, kk)
	case d1 == 0x5: // Skip next if vx == vy
		assm = fmt.Sprintf("SNE V%d, %d", x, kk)
	case d1 == 0x6: // set Vx = kk
		assm = fmt.Sprintf("LD V%d, %d", x, kk)
	case d1 == 0x7: // set Vx = vx + kk
		assm = fmt.Sprintf("ADD V%d, %d", x, kk)
	case d1 == 0x8 && d4 == 0x0: // vx = vy
		assm = fmt.Sprintf("LD V%d, V%d", x, y)
	case d1 == 0x8 && d4 == 0x1: // vx = vx | vy
		assm = fmt.Sprintf("OR V%d, V%d", x, y)
	case d1 == 0x8 && d4 == 0x2: // vx = vx & vy
		assm = fmt.Sprintf("AND V%d, V%d", x, y)
	case d1 == 0x8 && d4 == 0x3: // vx = vx xor vy
		assm = fmt.Sprintf("XOR V%d, V%d", x, y)
	case d1 == 0x8 && d4 == 0x4: // vx = vx + vy, vf = carry
		assm = fmt.Sprintf("ADD V%d, V%d", x, y)
	case d1 == 0x8 && d4 == 0x5: // vx = vx - vy, vf = not borrow
		assm = fmt.Sprintf("SUB V%d, V%d", x, y)
	case d1 == 0x8 && d4 == 0x6: // vx = vx shr 1
		assm = fmt.Sprintf("SHR V%d", x)
	case d1 == 0x8 && d4 == 0x7: // vx = vy - vx, vf = not borrow
		assm = fmt.Sprintf("SUBN V%d, V%d", x, y)
	case d1 == 0x8 && d4 == 0xE: // vx = vx shl 1
		assm = fmt.Sprintf("SHL V%d", x)
	case d1 == 0x9:
		assm = fmt.Sprintf("SNE V%d, V%d", x, y)
	case d1 == 0xa:
		assm = fmt.Sprintf("LD I, %#03x", nnn)
	case d1 == 0xb:
		assm = fmt.Sprintf("JP V0, %#03x", nnn)
	case d1 == 0xc:
		assm = fmt.Sprintf("RND V%d, %#02x", x, kk)
	case d1 == 0xd:
		assm = fmt.Sprintf("SHL V%d", x)
	case d1 == 0xe && d4 == 0xe:
		assm = fmt.Sprintf("SKP V%d", x)
	case d1 == 0xe && d4 == 0x1:
		assm = fmt.Sprintf("SKNP V%d", x)
	case d1 == 0xf && d4 == 0x7:
		assm = fmt.Sprintf("LD V%d, DT", x)
	case d1 == 0xf && d4 == 0xa:
		assm = fmt.Sprintf("LD V%d K", x)
	case d1 == 0xf && d3 == 0x1 && d4 == 0x5:
		assm = fmt.Sprintf("LD DT, V%d", x)
	case d1 == 0xf && d4 == 0x8:
		assm = fmt.Sprintf("LD ST, V%d", x)
	case d1 == 0xf && d4 == 0xe:
		assm = fmt.Sprintf("ADD I, V%d", x)
	case d1 == 0xf && d4 == 0x9:
		assm = fmt.Sprintf("LD F, V%d", x)
	case d1 == 0xf && d4 == 0x3:
		assm = fmt.Sprintf("LD B, V%d", x)
	case d1 == 0xf && d3 == 0x5 && d4 == 0x5:
		assm = fmt.Sprintf("LD [I], V%d", x)
	case d1 == 0xf && d3 == 0x6 && d4 == 0x5:
		assm = fmt.Sprintf("LD V%d, [I]", x)
	default:
		assm = fmt.Sprintf("RAW %#04x", instr)

	}
	return assm
}
