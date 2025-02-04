package chip8

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"time"
	// "math/bits"
)

// shouldn't break things
const width = 64
const height = 32
const pixels = width * height

// May break things
const ram = 4096
const stack = 16
const keys = 16
const reg = 16

const pcInit = 512

type Sys struct {

	// CHIP8 Spec
	vn         [reg]uint8
	i, pc      uint16
	sp, dt, st uint8

	ram   [ram]uint8
	stack [stack]uint16
	gfx   [pixels]uint8
	keys  [keys]uint8

	// Channels

	Height uint8
	Width  uint8

	GFXCh chan [pixels]uint8
	KeyCh chan [keys]uint8

	// Additional
	sop, eop uint16

	tick *time.Ticker
}

// Eventually add the start of program adress as a parameter to increase compatibility
func InitSys() *Sys {
	s := new(Sys)
	s.pc = pcInit
	s.sop = pcInit

	s.GFXCh = make(chan [pixels]uint8)
	s.KeyCh = make(chan [keys]uint8)

	s.Height = height
	s.Width = width
	return s
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

// Load the ROM into system memory
func (s *Sys) LoadRom(r io.Reader) {

	data, err := io.ReadAll(r)
	if err != nil {
		log.Println(err)
	}

	if len(data) > len(s.ram)-int(s.sop) {
		log.Println("ROM too big")
	}

	if len(data) == 0 {
		log.Println("Load attempt of zero-length ROM")
	}

	n := copy(s.ram[s.sop:], data)
	s.eop = uint16(n + int(s.sop))
	log.Println("loaded ", n, "bytes")

}

// Execute a single clock cycle
func (s *Sys) Step() {
	instr := uint16(s.ram[s.pc])<<8 | uint16(s.ram[s.pc])
	s.execute(instr)
}

// Run indefinetly
func (s *Sys) Run() {
	s.tick = time.NewTicker(16 * time.Millisecond)
	go s.keysDaemon()

	go func() {
		for {

			instr := uint16(s.ram[s.pc])<<8 | uint16(s.ram[s.pc])
			s.execute(instr)

			s.GFXCh <- s.gfx

			<-s.tick.C
		}

	}()

}

// Halt system
func (s *Sys) Halt() {
	s.tick.Stop()
}

func (s *Sys) Close() {
	close(s.GFXCh)
	close(s.KeyCh)
}

func (s *Sys) keysDaemon() {
	for {
		s.keys = <-s.KeyCh
	}
}

// Executes the provided instruction and increments SP
// INCOMPLETE
func (s *Sys) execute(instr uint16) {
	ret := true

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
		ret = false
	case d1 == 0x1: // jump to addr nnn
		s.pc = nnn
		ret = false
	case d1 == 0x2: // call addr nnn
		s.sp++
		s.stack[s.sp] = nnn
		ret = false
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

	if ret {
		s.pc += 2
	}
}

// Returns a readable assembly
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
