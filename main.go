package main

import (
	"image/color"
	"log"

	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
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
	romBtn := widget.NewButton("Load Rom", func() {
		romContainer.Objects = nil
		out := p.cpu.Disasm()

		for i := range out {
			//fmt.Println(out[i])
			romContainer.Objects = append(romContainer.Objects, widget.NewLabel(out[i]))

		}
		romContainer.Refresh()
	})

	runBtn := widget.NewButton("Run Emulator", func() { p.cpu.Run() })

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

	disp := p.buildDisplay(64, 32)

	bPanel := container.New(layout.NewGridLayout(2), romBtn, runBtn, openFile)

	content := container.New(layout.NewHBoxLayout(), bPanel, disp, romScroll)

	p.w.SetContent(content)
}

func (p *prog) buildDisplay(h int, w int) fyne.CanvasObject {
	// NewRaster(generate func(w,h int) image.Image) *Raster
	// NewImgeFromResource(res fyne.Resource) *Image
	//

	disp := canvas.NewRasterFromImage((p.frameGen(64, 32)))
	disp.SetMinSize(fyne.NewSize(64, 32))

	disp.ScaleMode = canvas.ImageScaleFastest

	return disp
}

func (p *prog) frameGen(w, h int) image.Image {
	img := image.NewGray(image.Rect(0, 0, w, h))

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			shade := uint8((x ^ y) % 256) // Simple pattern
			img.SetGray(x, y, color.Gray{Y: shade})
		}
	}

	return img
}
