package rendering

import (
	"bytes"
	_ "embed"
	"image/color"
	"log"
	"math"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/math/f32"

	"github.com/you/trajectory/constants"
	"github.com/you/trajectory/space/colors"
	Models "github.com/you/trajectory/space/model"
	"github.com/you/trajectory/space/shadows"
)

//----------------------------------------------------------------
// Shader embedding
//----------------------------------------------------------------

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

//go:embed blackhole.go
var blackHoleShader []byte

//go:embed whitehole.go
var whiteHoleShader []byte

//go:embed spiral_distortion.go
var spiralDistortionShader []byte

//go:embed portal_distortion.go
var portalDistortionShader []byte

//----------------------------------------------------------------
// Text rendering setup
//----------------------------------------------------------------

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

//----------------------------------------------------------------
// Shader cache (singleton-style), to avoid recompilation
//----------------------------------------------------------------

const (
	fovLight      = 90.0       // default field of view for light source
	baseArm       = float32(8) // half-length of each arm by default
	stretchFactor = float32(1) // how much longer the interior arms get
	lineWidth     = float32(2.25)
)

var (
	shaderCacheMu sync.Mutex
	shaderCache   = make(map[string]*ebiten.Shader)
)

func getShader(key string, src []byte) *ebiten.Shader {
	shaderCacheMu.Lock()
	defer shaderCacheMu.Unlock()
	if s, ok := shaderCache[key]; ok {
		// Uncomment for debugging: log.Printf("Retrieved shader %s from cache", key)
		return s
	}
	// Uncomment for debugging: log.Printf("Compiling shader %s", key)
	s, err := ebiten.NewShader(src)
	if err != nil {
		log.Fatalf("failed to compile shader %s: %v", key, err)
	}
	shaderCache[key] = s
	return s
}

//----------------------------------------------------------------
// Renderer definition
//----------------------------------------------------------------

type Renderer struct {
	shadowSystem           *shadows.ShadowSystem
	orbitShader            *ebiten.Shader
	playerTrailShader      *ebiten.Shader
	trajectoryArrowShader  *ebiten.Shader
	nebulaShader           *ebiten.Shader
	invertOnLightShader    *ebiten.Shader
	revealOnLightShader    *ebiten.Shader
	alienTrailShader       *ebiten.Shader
	planetShader           *ebiten.Shader
	blackHoleShader        *ebiten.Shader
	whiteHoleShader        *ebiten.Shader
	spiralDistortionShader *ebiten.Shader
	portalDistortionShader *ebiten.Shader

	whiteTexture       *ebiten.Image
	intermediateBuffer *ebiten.Image
	spiralBuffer       *ebiten.Image
	portalBuffer1      *ebiten.Image
	portalBuffer2      *ebiten.Image

	startTime time.Time
}

// NewRenderer constructs a Renderer, compiling each shader once.
func NewRenderer() *Renderer {
	// Create persistent images
	w, h := constants.ScreenWidth, constants.ScreenHeight
	whiteTex := ebiten.NewImage(w, h)
	whiteTex.Fill(color.White)

	return &Renderer{
		shadowSystem:           shadows.NewShadowSystem(w, h),
		orbitShader:            getShader("orbitLight", orbitLightShader),
		playerTrailShader:      getShader("playerTrail", playerTrailShader),
		trajectoryArrowShader:  getShader("trajectoryArrow", trajectoryArrowShader),
		nebulaShader:           getShader("nebulaBackground", nebulaBackgroundShader),
		invertOnLightShader:    getShader("invertOnLight", invertOnLightShader),
		revealOnLightShader:    getShader("revealOnLight", revealOnLightShader),
		alienTrailShader:       getShader("alienTrail", alienTrailShader),
		planetShader:           getShader("planet", planetShader),
		blackHoleShader:        getShader("blackhole", blackHoleShader),
		whiteHoleShader:        getShader("whitehole", whiteHoleShader),
		spiralDistortionShader: getShader("spiralDistortion", spiralDistortionShader),
		portalDistortionShader: getShader("portalDistortion", portalDistortionShader),

		whiteTexture:       whiteTex,
		intermediateBuffer: ebiten.NewImage(w, h),
		spiralBuffer:       ebiten.NewImage(w, h),
		portalBuffer1:      ebiten.NewImage(w, h),
		portalBuffer2:      ebiten.NewImage(w, h),

		startTime: time.Now(),
	}
}

// InitializeShaders forces all shaders to be fully initialized by performing dummy renders
func (r *Renderer) InitializeShaders() {
	// Create a 1x1 white texture for dummy renders
	tempTexture := ebiten.NewImage(1, 1)
	tempTexture.Fill(color.White)
	defer tempTexture.Dispose()

	// Create a minimal target image for dummy renders
	tempImage := ebiten.NewImage(1, 1)
	defer tempImage.Dispose()

	// Define basic dummy uniforms that should work with most shaders
	dummyUniforms := map[string]interface{}{
		"Time":           float32(0),
		"ScreenSize":     []float32{1, 1},
		"PlayerPos":      []float32{0, 0},
		"CameraPos":      []float32{0, 0},
		"Zoom":           float32(1),
		"Radius":         float32(1),
		"Velocity":       []float32{0, 0},
		"PlayerColor":    []float32{1, 1, 1, 1},
		"DropCount":      int(1),
		"TrailLength":    float32(1),
		"DropSizeMin":    float32(0.1),
		"DropSizeMax":    float32(0.2),
		"JitterAmt":      float32(0),
		"SpawnRate":      float32(1),
		"Lifetime":       float32(1),
		"OriginalColor":  []float32{1, 1, 1, 1},
		"LightPos":       []float32{0, 0},
		"LightDirection": []float32{0, 0},
		"FOVAngle":       float32(0),
		"MaxDistance":    float32(0),
		"NumBlackHoles":  int(0),
		"BHPositions":    []float32{0, 0, 0, 0, 0, 0},
		"OrbitRadii":     []float32{1, 1, 1},
		"Strengths":      []float32{1, 1, 1},
		"Portal_Pos":     []float32{0, 0},
		"Portal_Radius":  float32(1),
		"Portal_Color":   []float32{1, 1, 1},
		"IsActive":       float32(1),
	}

	// List of all shaders to initialize
	shaders := []*ebiten.Shader{
		r.orbitShader,
		r.playerTrailShader,
		r.trajectoryArrowShader,
		r.nebulaShader,
		r.invertOnLightShader,
		r.revealOnLightShader,
		r.alienTrailShader,
		r.planetShader,
		r.blackHoleShader,
		r.whiteHoleShader,
		r.spiralDistortionShader,
		r.portalDistortionShader,
	}

	// Perform dummy render for each shader
	for _, shader := range shaders {
		if shader != nil {
			opts := &ebiten.DrawRectShaderOptions{
				Uniforms: dummyUniforms,
			}
			opts.Images[0] = tempTexture
			tempImage.DrawRectShader(1, 1, shader, opts)
		}
	}
}

// Draw renders the full scene with post-processing effects.
func (r *Renderer) Draw(screen *ebiten.Image, model *Models.SpaceGame) {
	if r.spiralDistortionShader == nil {
		return
	}

	r.intermediateBuffer.Clear()
	r.drawNebulaBackground(r.intermediateBuffer, model)
	r.renderShadows(r.intermediateBuffer, model)
	r.drawCelestialBodies(r.intermediateBuffer, model)
	r.drawAsteroids(r.intermediateBuffer, model)
	r.drawPortals(r.intermediateBuffer, model)
	r.drawPlayerTrail(r.intermediateBuffer, model)
	r.drawPlayer(r.intermediateBuffer, model)
	r.drawTrajectoryArrow(r.intermediateBuffer, model)
	r.drawLastCollisionMarker(r.intermediateBuffer, model)
	r.drawBorderIndicators(r.intermediateBuffer, model)
	r.drawFPS(r.intermediateBuffer)

	r.spiralBuffer.Clear()
	r.applySpiralOverlay(r.spiralBuffer, r.intermediateBuffer, model, r.getBlackHoles(model))
	r.applyPortalDistortion(screen, r.spiralBuffer, model)
	r.DrawBlackHoles(screen, model)
}

// (Implement drawDirect, drawNebulaBackground, renderShadows, etc. exactly
// as in your original code, unchanged.)
// For brevity, those methods are omitted here but should be copied verbatim.

// getBlackHoles extracts black holes from the celestial bodies
func (r *Renderer) getBlackHoles(model *Models.SpaceGame) []*Models.BlackHole {
	var blackHoles []*Models.BlackHole
	for _, body := range model.CelestialBodies {
		if blackHole, ok := body.(*Models.BlackHole); ok {
			blackHoles = append(blackHoles, blackHole)
		}
	}
	return blackHoles
}

func (r *Renderer) applySpiralOverlay(
	screen *ebiten.Image,
	source *ebiten.Image,
	model *Models.SpaceGame,
	blackHoles []*Models.BlackHole,
) {
	// If there are black holes, apply distortion for up to 3 of them
	if len(blackHoles) > 0 {
		// elapsed time
		t := float32(time.Since(r.startTime).Seconds())

		// Determine how many black holes to process (max 3)
		numBH := len(blackHoles)
		if numBH > 3 {
			numBH = 3
		}

		// Prepare arrays for shader uniforms
		bhPositions := make([]float32, 6) // x1,y1, x2,y2, x3,y3
		orbitRadii := make([]float32, 3)  // radius1, radius2, radius3
		strengths := make([]float32, 3)   // strength1, strength2, strength3

		// Fill arrays with black hole data
		for i := 0; i < numBH; i++ {
			bh := blackHoles[i]
			c := model.Camera.WorldToScreen(bh.GetPosition(), constants.ScreenWidth, constants.ScreenHeight)
			orbitRadius := model.Camera.RadiusToScreen(bh.GetOrbitRadius(), constants.ScreenWidth, constants.ScreenHeight)

			// Store position as x,y pairs in flattened array
			bhPositions[i*2] = c[0]   // x
			bhPositions[i*2+1] = c[1] // y

			orbitRadii[i] = orbitRadius
			strengths[i] = float32(0.68) // strength for each black hole
		}

		uniforms := map[string]any{
			"NumBlackHoles": numBH,
			"Time":          t,
			"BHPositions":   bhPositions,
			"OrbitRadii":    orbitRadii,
			"Strengths":     strengths,
		}

		opts := &ebiten.DrawRectShaderOptions{
			Uniforms: uniforms,
		}
		opts.Images[0] = source

		// draw full-screen quad with our pixel-mode shader
		screen.DrawRectShader(constants.ScreenWidth, constants.ScreenHeight, r.spiralDistortionShader, opts)
	} else {
		// No black holes, just draw the source directly
		screen.DrawImage(source, nil)
	}
}

func (r *Renderer) applyPortalDistortion(
	screen *ebiten.Image,
	source *ebiten.Image,
	model *Models.SpaceGame,
) {
	// Skip if no portal distortion shader or no portals
	if r.portalDistortionShader == nil || len(model.Portals) == 0 {
		screen.DrawImage(source, nil)
		return
	}

	camera := model.Camera

	// Portal colors (different colors for different pair IDs)
	portalColors := []color.RGBA{
		{R: 0, G: 204, B: 255, A: 180},   // Cyan
		{R: 255, G: 102, B: 204, A: 180}, // Pink
		{R: 204, G: 255, B: 51, A: 180},  // Lime
		{R: 255, G: 153, B: 0, A: 180},   // Orange
		{R: 153, G: 51, B: 255, A: 180},  // Purple
		{R: 255, G: 255, B: 51, A: 180},  // Yellow
	}

	// Get current time for animations
	t := float32(time.Since(r.startTime).Seconds())

	// Use ping-pong buffers for multiple portal distortions
	var currentSource *ebiten.Image = source
	var currentTarget *ebiten.Image

	// Apply distortion for each portal sequentially
	for i, portal := range model.Portals {
		// Select color based on pair ID
		colorIndex := portal.PairID % len(portalColors)
		portalColor := portalColors[colorIndex]

		// Use the average of width and height as the circle radius
		portalRadius := portal.GetOrbitRadius()

		// Calculate activity factor (1.0 if active, 0.0 if cooldown)
		isActive := float32(1.0)
		if !portal.IsActive || portal.CooldownTimer > 0 {
			isActive = 0.0
		}

		// Set up shader uniforms
		uniforms := map[string]any{
			"Portal_Pos":    []float32{portal.Position[0], portal.Position[1]},
			"Portal_Radius": portalRadius,
			"Portal_Color":  []float32{float32(portalColor.R) / 255, float32(portalColor.G) / 255, float32(portalColor.B) / 255},
			"CameraPos":     []float32{camera.Position[0], camera.Position[1]},
			"Zoom":          camera.GetTotalZoom(),
			"Time":          t,
			"ScreenSize":    []float32{float32(constants.ScreenWidth), float32(constants.ScreenHeight)},
			"IsActive":      isActive,
		}

		// Set up shader options
		opts := &ebiten.DrawRectShaderOptions{
			Uniforms: uniforms,
		}
		opts.Images[0] = currentSource

		// For the last portal, draw directly to screen
		if i == len(model.Portals)-1 {
			screen.DrawRectShader(constants.ScreenWidth, constants.ScreenHeight, r.portalDistortionShader, opts)
		} else {
			// Use ping-pong buffers for intermediate portals
			if i%2 == 0 {
				currentTarget = r.portalBuffer1
			} else {
				currentTarget = r.portalBuffer2
			}

			// Clear the target buffer and apply distortion
			currentTarget.Clear()
			currentTarget.DrawRectShader(constants.ScreenWidth, constants.ScreenHeight, r.portalDistortionShader, opts)
			currentSource = currentTarget
		}
	}
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

	// Collect portals in screen coordinates
	var portalPositions []f32.Vec2
	var portalRadii []float32
	for _, portal := range model.Portals {
		portalPositions = append(portalPositions, camera.WorldToScreen(portal.GetPosition(), constants.ScreenWidth, constants.ScreenHeight))
		portalRadii = append(portalRadii, camera.RadiusToScreen(portal.GetRadius(), constants.ScreenWidth, constants.ScreenHeight))
	}

	// Collect portals rotations
	var portalRotations []float32
	for _, portal := range model.Portals {
		// Use the portal's rotation for distortion effects
		portalRotations = append(portalRotations, portal.Rotation)
	}

	// Use player position as light source in screen coordinates
	lightPos := camera.WorldToScreen(model.Player.Position, constants.ScreenWidth, constants.ScreenHeight)
	lightDirection := camera.WorldToScreen(camera.Position, constants.ScreenWidth, constants.ScreenHeight)
	fov := r.getAdaptiveFov(lightDirection, lightPos)

	r.shadowSystem.RenderShadows(screen, lightPos, lightDirection, fov, camera.GetTotalZoom(), celestialPositions, celestialRadii, asteroidPositions, asteroidRadii, portalPositions, portalRadii, portalRotations, false)
}

// drawTrajectoryArrow draws the trajectory arrow when dragging
func (r *Renderer) drawTrajectoryArrow(screen *ebiten.Image, model *Models.SpaceGame) {
	// This would need access to input dragInfo - will be handled in the main game loop
	// For now, this is a placeholder for the trajectory arrow rendering logic
}

// drawLastCollisionMarker draws red "x" markers at all collision positions from history
func (r *Renderer) drawLastCollisionMarker(screen *ebiten.Image, model *Models.SpaceGame) {
	if len(model.CollisionHistory) == 0 {
		return
	}

	camera := model.Camera

	// Draw all collision markers, with different sizes/transparency based on age
	for i, collision := range model.CollisionHistory {
		screenPos := camera.WorldToScreen(collision.Position, constants.ScreenWidth, constants.ScreenHeight)

		// Scale size and alpha based on age (newest = largest/most opaque)
		ageFactor := 1.0 - float32(i)*0.2 // 1.0, 0.7, 0.4 for positions 0, 1, 2
		if ageFactor < 0.3 {
			ageFactor = 0.3 // Minimum visibility
		}

		crossSize := float32(10.0) * ageFactor // Half-size of the "x" in pixels (diagonal span)
		lineWidth := float32(6.0) * ageFactor

		// Create color with reduced alpha for older collisions
		crossColor := color.RGBA{R: colors.LastCollisionCrossColor.R, G: colors.LastCollisionCrossColor.G, B: colors.LastCollisionCrossColor.B, A: 255}

		// Draw two diagonal lines to form an "x"
		// Top-left to bottom-right
		r.drawLine(screen,
			screenPos[0]-crossSize, screenPos[1]-crossSize,
			screenPos[0]+crossSize, screenPos[1]+crossSize,
			lineWidth, crossColor)

		// Top-right to bottom-left
		r.drawLine(screen,
			screenPos[0]+crossSize, screenPos[1]-crossSize,
			screenPos[0]-crossSize, screenPos[1]+crossSize,
			lineWidth, crossColor)
	}
}

// drawLine draws a line between two points
func (r *Renderer) drawLine(screen *ebiten.Image, x1, y1, x2, y2, width float32, color color.RGBA) {
	// Calculate line vector
	dx := x2 - x1
	dy := y2 - y1
	length := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	if length == 0 {
		return
	}

	// Normalize vector
	dx /= length
	dy /= length

	// Perpendicular vector for width
	px := -dy * width / 2
	py := dx * width / 2

	// Draw line as a quad
	vertices := []ebiten.Vertex{
		{DstX: x1 + px, DstY: y1 + py, SrcX: 0, SrcY: 0, ColorR: float32(color.R) / 255, ColorG: float32(color.G) / 255, ColorB: float32(color.B) / 255, ColorA: float32(color.A) / 255},
		{DstX: x1 - px, DstY: y1 - py, SrcX: 0, SrcY: 0, ColorR: float32(color.R) / 255, ColorG: float32(color.G) / 255, ColorB: float32(color.B) / 255, ColorA: float32(color.A) / 255},
		{DstX: x2 - px, DstY: y2 - py, SrcX: 0, SrcY: 0, ColorR: float32(color.R) / 255, ColorG: float32(color.G) / 255, ColorB: float32(color.B) / 255, ColorA: float32(color.A) / 255},
		{DstX: x2 + px, DstY: y2 + py, SrcX: 0, SrcY: 0, ColorR: float32(color.R) / 255, ColorG: float32(color.G) / 255, ColorB: float32(color.B) / 255, ColorA: float32(color.A) / 255},
	}

	indices := []uint16{0, 1, 2, 0, 2, 3}

	screen.DrawTriangles(vertices, indices, r.whiteTexture, nil)
}

// drawPortals renders all portals in the game independently
func (r *Renderer) drawPortals(screen *ebiten.Image, model *Models.SpaceGame) {
	for _, portal := range model.Portals {
		r.drawPortalRing(screen, portal, model)
	}
}

// drawPortalRing renders a single portal as a stroked ring
func (r *Renderer) drawPortalRing(screen *ebiten.Image, portal *Models.Portal, model *Models.SpaceGame) {
	camera := model.Camera

	// Convert portal position to screen coordinates
	screenPos := camera.WorldToScreen(portal.Position, constants.ScreenWidth, constants.ScreenHeight)

	// Compute radius in screen pixels
	portalRadius := camera.RadiusToScreen(portal.GetRadius(), constants.ScreenWidth, constants.ScreenHeight)

	// Choose color by PairID
	portalColors := colors.PortalColors
	clr := portalColors[portal.PairID%len(portalColors)]
	if !portal.IsActive || portal.CooldownTimer > 0 {
		clr = colors.PortalColorInactive // Use inactive color if not active
	}

	ringWidth := portalRadius * 0.1
	radiusInner := portalRadius - ringWidth

	// Draw outer circle
	vector.DrawFilledCircle(screen, screenPos[0], screenPos[1], portalRadius, clr, true)
	// Draw inner circle
	vector.DrawFilledCircle(screen, screenPos[0], screenPos[1], radiusInner, colors.PortalColorInnerRing, true)
}
