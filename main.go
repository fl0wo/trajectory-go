package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/you/trajectory/gui"
	"log"
)

func main() {
	app, err := gui.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	w, h := ebiten.Monitor().Size()
	ebiten.SetWindowSize(int(float32(w)/1.5), int(float32(h)/1.5))
	ebiten.SetWindowTitle("Trajectory Space Game")
	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}
