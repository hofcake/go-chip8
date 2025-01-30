package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/hofcake/go-chip8/chip8"
)

func main() {
	fmt.Println("Hello, World!")

	sys := chip8.InitSys()

	sys.LoadRom("roms/Keypad Test [Hap, 2006].ch8")

	/*
		for i := 0; i < 10; i++ {
			sys.PrintNext()
		} */

	a := app.New()
	w := a.NewWindow("Hello World")
	//w2 := a.NewWindow("Second Window")

	romContainer := container.NewVBox()
	romScroll := container.NewVScroll(romContainer)

	romBtn := widget.NewButton("Load Rom", func() {
		displayRom(sys, romContainer)
	})

	content := container.New(layout.NewHBoxLayout(), romBtn, romScroll)

	w.SetContent(content)

	prependTo(romContainer, "test")

	w.Show()
	//w2.Show()
	a.Run()

	//sys.Dump()

}

func displayRom(s *chip8.Sys, g *fyne.Container) {
	remRom(g)
	out := s.Disasm()
	i := 0
	for range out {
		if out[i] != "" {
			//fmt.Println(out[i])
			prependTo(g, out[i])
		}
		i++
	}
}

func prependTo(g *fyne.Container, s string) {
	g.Objects = append([]fyne.CanvasObject{widget.NewLabel(s)}, g.Objects...)
	g.Refresh()
}

func remRom(g *fyne.Container) {
	g.Objects = nil
}
