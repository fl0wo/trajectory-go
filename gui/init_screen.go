package gui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/you/trajectory/space/rendering"
)

// InitScreen is a temporary screen to initialize all shaders at startup.
type InitScreen struct {
	screenManager *ScreenManager
	renderer      *rendering.Renderer
	initialized   bool
}

// NewInitScreen creates a new InitScreen instance.
func NewInitScreen(screenManager *ScreenManager) *InitScreen {
	return &InitScreen{
		screenManager: screenManager,
		renderer:      rendering.NewRenderer(),
		initialized:   false,
	}
}

// Update checks if initialization is done and switches to the home screen.
func (i *InitScreen) Update() error {
	if i.initialized {
		i.screenManager.SetScreen(HomeScreen)
	}
	return nil
}

// Draw performs a dummy render with all shaders to ensure they're initialized.
func (i *InitScreen) Draw(screen *ebiten.Image) {
	if !i.initialized {
		// Fill screen with a dark color during initialization
		screen.Fill(color.RGBA{20, 20, 30, 255})

		// Force initialization of all shaders
		i.renderer.InitializeShaders()

		i.initialized = true
	}
}

// Layout returns the screen dimensions.
func (i *InitScreen) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
