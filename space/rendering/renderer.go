package rendering

import (
	"bytes"
	_ "embed"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/you/trajectory/constants"
	Models "github.com/you/trajectory/space/model"
	"github.com/you/trajectory/space/shadows"
	"golang.org/x/image/math/f32"
	"image/color"
	"log"
	"time"
)

//go:embed orbit_light.go
var orbitLightShader []byte

//go:embed player_trail.go
var playerTrailShader []byte

//go:embed trajectory_arrow.go
var trajectoryArrowShader []byte

//go:embed nebula_background.go
var nebulaBackgroundShader []byte

//go:embed invert_on_light.go
var invertOnLightShader []byte

//go:embed reveal_on_light.go
var revealOnLightShader []byte

//go:embed alien_trail.go
var alienTrailShader []byte

//go:embed planet.go
var planetShader []byte

var (
	mplusFaceSource *text.GoTextFaceSource
)

func init() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal(err)
	}
	mplusFaceSource = s
}

const (
	fovLight = 90.0
)

const (
	baseArm       = float32(8) // half-length of each arm by default
	stretchFactor = float32(1) // how much longer the interior arms get
	lineWidth     = float32(2.25)
)

// Renderer handles all rendering operations for the game
type Renderer struct {
	shadowSystem          *shadows.ShadowSystem
	orbitShader           *ebiten.Shader
	playerTrailShader     *ebiten.Shader
	trajectoryArrowShader *ebiten.Shader
	nebulaShader          *ebiten.Shader
	invertOnLightShader   *ebiten.Shader
	revealOnLightShader   *ebiten.Shader
	alienTrailShader      *ebiten.Shader
	planetShader          *ebiten.Shader
	whiteTexture          *ebiten.Image
	startTime             time.Time // Game start time for shader animations
}

// NewRenderer creates a new renderer instance
func NewRenderer() *Renderer {
	// Initialize orbit shader
	var orbitShader *ebiten.Shader
	if shader, err := ebiten.NewShader(orbitLightShader); err == nil {
		orbitShader = shader
	}

	// Initialize player trail shader
	var playerTrailShaderObj *ebiten.Shader
	if shader, err := ebiten.NewShader(playerTrailShader); err == nil {
		playerTrailShaderObj = shader
	}

	// Initialize trajectory arrow shader
	var trajectoryArrowShaderObj *ebiten.Shader
	if shader, err := ebiten.NewShader(trajectoryArrowShader); err == nil {
		trajectoryArrowShaderObj = shader
	}

	// Initialize nebula background shader
	var nebulaShader *ebiten.Shader
	if shader, err := ebiten.NewShader(nebulaBackgroundShader); err == nil {
		nebulaShader = shader
	}

	// Initialize invert on light shader
	var invertOnLightShaderObj *ebiten.Shader
	if shader, err := ebiten.NewShader(invertOnLightShader); err == nil {
		invertOnLightShaderObj = shader
	}

	// Initialize reveal on light shader
	var revealOnLightShaderObj *ebiten.Shader
	if shader, err := ebiten.NewShader(revealOnLightShader); err == nil {
		revealOnLightShaderObj = shader
	}

	// Initialize alien trail shader
	var alienTrailShaderObj *ebiten.Shader
	if shader, err := ebiten.NewShader(alienTrailShader); err == nil {
		alienTrailShaderObj = shader
	} else {
		log.Println("Failed to load alien trail shader:", err)
	}

	// Initialize planet shader
	var planetShaderObj *ebiten.Shader
	if shader, err := ebiten.NewShader(planetShader); err == nil {
		planetShaderObj = shader
	} else {
		log.Println("Failed to load planet shader:", err)
	}

	// Create a white texture matching screen size for shaders that need a source image
	whiteTexture := ebiten.NewImage(constants.ScreenWidth, constants.ScreenHeight)
	whiteTexture.Fill(color.White)

	return &Renderer{
		shadowSystem:          shadows.NewShadowSystem(constants.ScreenWidth, constants.ScreenHeight),
		orbitShader:           orbitShader,
		playerTrailShader:     playerTrailShaderObj,
		trajectoryArrowShader: trajectoryArrowShaderObj,
		nebulaShader:          nebulaShader,
		invertOnLightShader:   invertOnLightShaderObj,
		revealOnLightShader:   revealOnLightShaderObj,
		alienTrailShader:      alienTrailShaderObj,
		planetShader:          planetShaderObj,
		whiteTexture:          whiteTexture,
		startTime:             time.Now(), // Initialize start time for shader animations
	}
}

// Draw renders the complete game scene
func (r *Renderer) Draw(screen *ebiten.Image, model *Models.SpaceGame) {
	// Draw dynamic nebula background
	r.drawNebulaBackground(screen, model)

	// Render shadows if enabled
	r.renderShadows(screen, model)

	// Draw celestial bodies
	r.drawCelestialBodies(screen, model)

	// Draw asteroids
	r.drawAsteroids(screen, model)

	// Draw player trail
	r.drawPlayerTrail(screen, model)

	// Draw player
	r.drawPlayer(screen, model)

	// Draw trajectory arrow if dragging
	r.drawTrajectoryArrow(screen, model)

	// Draw border indicators
	r.drawBorderIndicators(screen, model)

	// Draw FPS counter
	r.drawFPS(screen)
}

// renderShadows handles shadow rendering
func (r *Renderer) renderShadows(screen *ebiten.Image, model *Models.SpaceGame) {
	if !model.ShadowsEnabled || r.shadowSystem == nil {
		return
	}

	camera := model.Camera

	// Collect celestial bodies in screen coordinates
	var celestialPositions []f32.Vec2
	var celestialRadii []float32
	for _, body := range model.CelestialBodies {
		celestialPositions = append(celestialPositions, camera.WorldToScreen(body.GetPosition(), constants.ScreenWidth, constants.ScreenHeight))
		celestialRadii = append(celestialRadii, camera.RadiusToScreen(body.GetRadius(), constants.ScreenWidth, constants.ScreenHeight))
	}

	// Collect asteroids in screen coordinates
	var asteroidPositions []f32.Vec2
	var asteroidRadii []float32
	for _, asteroid := range model.RingAsteroids {
		asteroidPositions = append(asteroidPositions, camera.WorldToScreen(asteroid.GetPosition(), constants.ScreenWidth, constants.ScreenHeight))
		asteroidRadii = append(asteroidRadii, camera.RadiusToScreen(asteroid.GetRadius(), constants.ScreenWidth, constants.ScreenHeight))
	}

	// Use player position as light source in screen coordinates
	lightPos := camera.WorldToScreen(model.Player.Position, constants.ScreenWidth, constants.ScreenHeight)
	lightDirection := camera.WorldToScreen(camera.Position, constants.ScreenWidth, constants.ScreenHeight)
	fov := r.getAdaptiveFov(lightDirection, lightPos)

	r.shadowSystem.RenderShadows(screen, lightPos, lightDirection, fov, camera.GetTotalZoom(), celestialPositions, celestialRadii, asteroidPositions, asteroidRadii, false)
}

// drawTrajectoryArrow draws the trajectory arrow when dragging
func (r *Renderer) drawTrajectoryArrow(screen *ebiten.Image, model *Models.SpaceGame) {
	// This would need access to input dragInfo - will be handled in the main game loop
	// For now, this is a placeholder for the trajectory arrow rendering logic
}
