package Models

import "golang.org/x/image/math/f32"

type Planet struct {
	Name     string
	Position f32.Vec2
	Radius   float32
}
