package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/math/f32"
	"math"
)

// Dir represents a direction.
type Dir int

const (
	DirUp Dir = iota
	DirRight
	DirDown
	DirLeft
)

type mouseState int

const (
	mouseStateNone mouseState = iota
	mouseStateDragging
	mouseStateReleased
)

type touchState int

const (
	touchStateNone touchState = iota
	touchStatePressing
	touchStateSettled
	touchStateInvalid
)

const (
	scrollSensitivity = 10.0 // Adjust this value to change scroll sensitivity
)

// DragInfo contains information about the current drag operation
type DragInfo struct {
	IsDragging   bool
	StartPos     f32.Vec2
	CurrentPos   f32.Vec2
	DragVector   f32.Vec2 // Vector from start to current position
	DragDistance float32  // Length of drag vector
	IsReleased   bool     // True for one frame when drag is released
}

// String returns a string representing the direction.
func (d Dir) String() string {
	switch d {
	case DirUp:
		return "Up"
	case DirRight:
		return "Right"
	case DirDown:
		return "Down"
	case DirLeft:
		return "Left"
	}
	panic("not reach")
}

// Vector returns a [-1, 1] value for each axis.
func (d Dir) Vector() (x, y int) {
	switch d {
	case DirUp:
		return 0, -1
	case DirRight:
		return 1, 0
	case DirDown:
		return 0, 1
	case DirLeft:
		return -1, 0
	}
	panic("not reach")
}

// Input represents the current key states.
type Input struct {
	mouseState    mouseState
	mouseInitPosX int
	mouseInitPosY int
	mouseDir      Dir

	// Drag-related fields
	dragInfo DragInfo

	// Scroll-related fields
	scrollDelta float32 // Mouse wheel scroll delta

	touches       []ebiten.TouchID
	touchState    touchState
	touchID       ebiten.TouchID
	touchInitPosX int
	touchInitPosY int
	touchLastPosX int
	touchLastPosY int
	touchDir      Dir
}

// NewInput generates a new Input object.
func NewInput() *Input {
	return &Input{}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func vecToDir(dx, dy int) (Dir, bool) {
	if abs(dx) < 4 && abs(dy) < 4 {
		return 0, false
	}
	if abs(dx) < abs(dy) {
		if dy < 0 {
			return DirUp, true
		}
		return DirDown, true
	}
	if dx < 0 {
		return DirLeft, true
	}
	return DirRight, true
}

// Update updates the current input states.
func (i *Input) Update() {
	// Reset release flag
	i.dragInfo.IsReleased = false

	// Update scroll delta
	_, scrollY := ebiten.Wheel()
	i.scrollDelta = float32(scrollY) / scrollSensitivity // Adjust sensitivity as needed

	switch i.mouseState {
	case mouseStateNone:
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			i.mouseInitPosX = x
			i.mouseInitPosY = y
			i.mouseState = mouseStateDragging

			// Initialize drag info
			i.dragInfo.IsDragging = true
			i.dragInfo.StartPos = f32.Vec2{float32(x), float32(y)}
			i.dragInfo.CurrentPos = i.dragInfo.StartPos
			i.dragInfo.DragVector = f32.Vec2{0, 0}
			i.dragInfo.DragDistance = 0
		}
	case mouseStateDragging:
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			// Update current position and drag vector
			x, y := ebiten.CursorPosition()
			i.dragInfo.CurrentPos = f32.Vec2{float32(x), float32(y)}
			i.dragInfo.DragVector = f32.Vec2{
				i.dragInfo.CurrentPos[0] - i.dragInfo.StartPos[0],
				i.dragInfo.CurrentPos[1] - i.dragInfo.StartPos[1],
			}
			i.dragInfo.DragDistance = float32(math.Sqrt(float64(i.dragInfo.DragVector[0]*i.dragInfo.DragVector[0] + i.dragInfo.DragVector[1]*i.dragInfo.DragVector[1])))
		} else {
			// Mouse released - trigger throw
			i.dragInfo.IsReleased = true
			i.dragInfo.IsDragging = false
			i.mouseState = mouseStateReleased
		}
	case mouseStateReleased:
		i.mouseState = mouseStateNone
	}

	i.touches = ebiten.AppendTouchIDs(i.touches[:0])
	switch i.touchState {
	case touchStateNone:
		if len(i.touches) == 1 {
			i.touchID = i.touches[0]
			x, y := ebiten.TouchPosition(i.touches[0])
			i.touchInitPosX = x
			i.touchInitPosY = y
			i.touchLastPosX = x
			i.touchLastPosX = y
			i.touchState = touchStatePressing
		}
	case touchStatePressing:
		if len(i.touches) >= 2 {
			break
		}
		if len(i.touches) == 1 {
			if i.touches[0] != i.touchID {
				i.touchState = touchStateInvalid
			} else {
				x, y := ebiten.TouchPosition(i.touches[0])
				i.touchLastPosX = x
				i.touchLastPosY = y
			}
			break
		}
		if len(i.touches) == 0 {
			dx := i.touchLastPosX - i.touchInitPosX
			dy := i.touchLastPosY - i.touchInitPosY
			d, ok := vecToDir(dx, dy)
			if !ok {
				i.touchState = touchStateNone
				break
			}
			i.touchDir = d
			i.touchState = touchStateSettled
		}
	case touchStateSettled:
		i.touchState = touchStateNone
	case touchStateInvalid:
		if len(i.touches) == 0 {
			i.touchState = touchStateNone
		}
	}
}

// Dir returns a currently pressed direction.
// Dir returns false if no direction key is pressed.
func (i *Input) Dir() (Dir, bool) {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		return DirUp, true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		return DirLeft, true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		return DirRight, true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		return DirDown, true
	}
	if i.mouseState == mouseStateReleased {
		return i.mouseDir, true
	}
	if i.touchState == touchStateSettled {
		return i.touchDir, true
	}
	return 0, false
}

// GetDragInfo returns the current drag information
func (i *Input) GetDragInfo() DragInfo {
	return i.dragInfo
}

// GetScrollDelta returns the mouse wheel scroll delta for this frame
func (i *Input) GetScrollDelta() float32 {
	return i.scrollDelta
}
