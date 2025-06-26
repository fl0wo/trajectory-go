package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/you/trajectory/space"
	"log"
)

func main() {
	game, err := space.NewGame()
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowSize(1920, 1080) // Set to your desired resolution
	ebiten.SetWindowTitle("Trajectory Space Game")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
