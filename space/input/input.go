package input

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/math/f32"
)

// Dir represents a direction.
type Dir int

const (
	DirUp Dir = iota
	DirRight
	DirDown
	DirLeft
)

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

// DragInfo contains information about the current drag operation.
type DragInfo struct {
	IsDragging   bool
	StartPos     f32.Vec2
	CurrentPos   f32.Vec2
	DragVector   f32.Vec2 // Vector from start to current position
	DragDistance float32  // Length of drag vector
	IsReleased   bool     // True for one frame when drag is released
}

// Input represents the current key and pointer states.
type Input struct {
	// Mouse drag state
	mouseState    mouseState
	mouseInitPosX int
	mouseInitPosY int
	mouseDir      Dir

	// Drag info exposed
	dragInfo DragInfo

	// Scroll wheel delta this frame
	scrollDelta float32

	// Keyboard one-shot states
	restartKeyPressed   bool // R
	levelKeyPressed     int  // 1–9
	cameraTogglePressed bool // C
	shadowTogglePressed bool // S

	// Arrow keys (NEW)
	leftArrowPressed  bool // ← just pressed
	rightArrowPressed bool // → just pressed

	// Touch state
	touches       []ebiten.TouchID
	touchState    touchState
	touchID       ebiten.TouchID
	touchInitPosX int
	touchInitPosY int
	touchLastPosX int
	touchLastPosY int
	touchDir      Dir
}

// NewInput creates a new Input tracker.
func NewInput() *Input {
	return &Input{}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// vecToDir converts a delta to a Dir if the movement is large enough.
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

// Update polls all inputs and updates the internal state.
// Call this once per frame at the top of your game loop.
func (i *Input) Update() {
	// Reset one-shot flags
	i.dragInfo.IsReleased = false
	i.restartKeyPressed = false
	i.cameraTogglePressed = false
	i.shadowTogglePressed = false
	i.levelKeyPressed = 0
	i.leftArrowPressed = false
	i.rightArrowPressed = false

	// Scroll wheel
	_, scrollY := ebiten.Wheel()
	i.scrollDelta = float32(scrollY) / scrollSensitivity

	// Keyboard
	i.restartKeyPressed = inpututil.IsKeyJustPressed(ebiten.KeyR)
	i.cameraTogglePressed = inpututil.IsKeyJustPressed(ebiten.KeyC)
	i.shadowTogglePressed = inpututil.IsKeyJustPressed(ebiten.KeyS)

	// Level keys 1–9
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyDigit1):
		i.levelKeyPressed = 1
	case inpututil.IsKeyJustPressed(ebiten.KeyDigit2):
		i.levelKeyPressed = 2
	case inpututil.IsKeyJustPressed(ebiten.KeyDigit3):
		i.levelKeyPressed = 3
	case inpututil.IsKeyJustPressed(ebiten.KeyDigit4):
		i.levelKeyPressed = 4
	case inpututil.IsKeyJustPressed(ebiten.KeyDigit5):
		i.levelKeyPressed = 5
	case inpututil.IsKeyJustPressed(ebiten.KeyDigit6):
		i.levelKeyPressed = 6
	case inpututil.IsKeyJustPressed(ebiten.KeyDigit7):
		i.levelKeyPressed = 7
	case inpututil.IsKeyJustPressed(ebiten.KeyDigit8):
		i.levelKeyPressed = 8
	case inpututil.IsKeyJustPressed(ebiten.KeyDigit9):
		i.levelKeyPressed = 9
	}

	// Arrow keys (NEW)
	i.leftArrowPressed = inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft)
	i.rightArrowPressed = inpututil.IsKeyJustPressed(ebiten.KeyArrowRight)

	// Mouse drag handling
	switch i.mouseState {
	case mouseStateNone:
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			i.mouseInitPosX = x
			i.mouseInitPosY = y
			i.mouseState = mouseStateDragging

			i.dragInfo = DragInfo{
				IsDragging:   true,
				StartPos:     f32.Vec2{float32(x), float32(y)},
				CurrentPos:   f32.Vec2{float32(x), float32(y)},
				DragVector:   f32.Vec2{0, 0},
				DragDistance: 0,
				IsReleased:   false,
			}
		}

	case mouseStateDragging:
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			i.dragInfo.CurrentPos = f32.Vec2{float32(x), float32(y)}
			dx := i.dragInfo.CurrentPos[0] - i.dragInfo.StartPos[0]
			dy := i.dragInfo.CurrentPos[1] - i.dragInfo.StartPos[1]
			i.dragInfo.DragVector = f32.Vec2{dx, dy}
			i.dragInfo.DragDistance = float32(math.Hypot(float64(dx), float64(dy)))
		} else {
			// Release
			i.dragInfo.IsReleased = true
			i.dragInfo.IsDragging = false
			i.mouseState = mouseStateReleased
		}

	case mouseStateReleased:
		i.mouseState = mouseStateNone
	}

	// Touch handling
	i.touches = ebiten.AppendTouchIDs(i.touches[:0])
	switch i.touchState {
	case touchStateNone:
		if len(i.touches) == 1 {
			i.touchID = i.touches[0]
			x, y := ebiten.TouchPosition(i.touchID)
			i.touchInitPosX = x
			i.touchInitPosY = y
			i.touchLastPosX = x
			i.touchLastPosY = y
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
				x, y := ebiten.TouchPosition(i.touchID)
				i.touchLastPosX = x
				i.touchLastPosY = y
			}
			break
		}
		if len(i.touches) == 0 {
			dx := i.touchLastPosX - i.touchInitPosX
			dy := i.touchLastPosY - i.touchInitPosY
			if d, ok := vecToDir(dx, dy); ok {
				i.touchDir = d
				i.touchState = touchStateSettled
			} else {
				i.touchState = touchStateNone
			}
		}

	case touchStateSettled:
		i.touchState = touchStateNone

	case touchStateInvalid:
		if len(i.touches) == 0 {
			i.touchState = touchStateNone
		}
	}
}

// Dir returns a just-pressed direction (arrow, mouse drag, or touch swipe).
// The boolean is true if a direction input was detected this frame.
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

// GetDragInfo returns the current drag information.
func (i *Input) GetDragInfo() DragInfo {
	return i.dragInfo
}

// GetScrollDelta returns the mouse wheel scroll delta for this frame.
func (i *Input) GetScrollDelta() float32 {
	return i.scrollDelta
}

// IsRestartPressed returns true if the R key was just pressed this frame.
func (i *Input) IsRestartPressed() bool {
	return i.restartKeyPressed
}

// GetLevelKeyPressed returns the level number (1-9) if a level key was pressed, 0 otherwise.
func (i *Input) GetLevelKeyPressed() int {
	return i.levelKeyPressed
}

// IsCameraTogglePressed returns true if the C key was just pressed this frame.
func (i *Input) IsCameraTogglePressed() bool {
	return i.cameraTogglePressed
}

// IsShadowTogglePressed returns true if the S key was just pressed this frame.
func (i *Input) IsShadowTogglePressed() bool {
	return i.shadowTogglePressed
}

// IsLeftArrowPressed returns true if the Left arrow was just pressed this frame.
func (i *Input) IsLeftArrowPressed() bool {
	return i.leftArrowPressed
}

// IsRightArrowPressed returns true if the Right arrow was just pressed this frame.
func (i *Input) IsRightArrowPressed() bool {
	return i.rightArrowPressed
}
