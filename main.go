package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/hofcake/go-chip8/chip8"
)

type prog struct {
	cpu *chip8.Sys
	a   fyne.App
	w   fyne.Window
}

func initProg() *prog {
	p := new(prog)

	p.cpu = chip8.InitSys()
	p.a = app.New()
	p.w = p.a.NewWindow("Go-Chip8 Emulator")

	return p

}

func main() {

	p := initProg()
	p.buildUI()

	//prog.w.Show()
	//prog.a.Run()
	p.w.ShowAndRun()

}

func (p *prog) buildUI() {
	// Assembly view
	romContainer := container.NewVBox()
	romScroll := container.NewVScroll(romContainer)

	// control panel
	/*
		romBtn := widget.NewButton("Load Rom", func() {
			displayRom(romContainer)
		})*/

	romBtn := widget.NewButton("Load Rom", func() {
		romContainer.Objects = nil
		out := p.cpu.Disasm()

		for i := range out {
			//fmt.Println(out[i])
			romContainer.Objects = append(romContainer.Objects, widget.NewLabel(out[i]))

		}
		romContainer.Refresh()
	})

	// File Dialog
	openFile := widget.NewButton("Load Rom (.ch8)", func() {
		fd := dialog.NewFileOpen(
			func(r fyne.URIReadCloser, err error) {
				if err != nil {
					log.Println("Error with file open dialog")
				}
				p.cpu.LoadRom(r)
			}, p.w)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".ch8"}))
		fd.Show()
	})

	content := container.New(layout.NewHBoxLayout(), romBtn, openFile, romScroll)

	p.w.SetContent(content)
}
