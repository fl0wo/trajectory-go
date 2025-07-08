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

		// now subtracted 10sec
		startTime: time.Now().Add(-10 * time.Second),
	}
}

// Draw renders the full scene with post-processing effects.
func (r *Renderer) Draw(screen *ebiten.Image, model *Models.SpaceGame) {
	if r.spiralDistortionShader == nil {
		return
	}

	// Draw all game objects directly on screen
	r.drawNebulaBackground(screen, model)
	r.renderShadows(screen, model)
	r.drawCelestialBodies(screen, model)
	r.drawAsteroids(screen, model)
	r.drawPortals(screen, model)
	r.drawPlayerTrail(screen, model)
	r.drawPlayer(screen, model)
	r.drawTrajectoryArrow(screen, model)
	r.drawLastCollisionMarker(screen, model)
	r.drawBorderIndicators(screen, model)
	r.drawFPS(screen)
}

func (r *Renderer) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, geoM ebiten.GeoM, model *Models.SpaceGame) {
	// Apply full-screen effects if needed
	if len(model.Portals) > 0 {
		// uses intermitted buffer to bridge from offscreen to offscreen
		r.applyPortalDistortion(offscreen, offscreen, model, geoM)
	}

	hasBlackholes := r.applySpiralOverlay(screen, offscreen, model, geoM)
	if hasBlackholes {
		// r.DrawBlackHoles(offscreen, model)
	} else {
		// If no black holes, just draw the offscreen image directly
		screen.DrawImage(offscreen, &ebiten.DrawImageOptions{GeoM: geoM})
	}
}

func (r *Renderer) applySpiralOverlay(
	screen ebiten.FinalScreen,
	offscreen *ebiten.Image,
	model *Models.SpaceGame,
	geoM ebiten.GeoM,
) bool {

	if r.spiralDistortionShader == nil {
		return false
	}

	t := float32(time.Since(r.startTime).Seconds())

	bhPositions := make([]float32, 8)
	orbitRadii := make([]float32, 4)
	strengths := make([]float32, 4)

	numBH := 0
	// First pass: draw black hole orbits
	for i, body := range model.CelestialBodies {
		// Only render black holes
		if body.GetType() != Models.CelestialBodyTypeBlackHole {
			continue
		}

		bh := body.(*Models.BlackHole)
		c := model.Camera.WorldToScreen(bh.GetPosition(), constants.ScreenWidth, constants.ScreenHeight)
		orbitRadius := model.Camera.RadiusToScreen(bh.GetOrbitRadius(), constants.ScreenWidth, constants.ScreenHeight)
		bhPositions[i*2] = c[0]
		bhPositions[i*2+1] = c[1]
		orbitRadii[i] = orbitRadius
		strengths[i] = 0.68
		numBH++
	}

	if numBH == 0 {
		// No black holes, skip shader application
		return false
	}

	if numBH > 4 {
		print("Warning: More than 4 black holes detected, limiting to 4.\n")
		numBH = 4 // Limit to 4 black holes
	}

	uniforms := map[string]any{
		"NumBlackHoles": numBH,
		"Time":          t,
		"BHPositions":   bhPositions,
		"OrbitRadii":    orbitRadii,
		"Strengths":     strengths,

		"ScreenSize": []float32{float32(constants.ScreenWidth), float32(constants.ScreenHeight)},
		"Zoom":       model.Camera.GetTotalZoom(),
		"CameraPos":  []float32{model.Camera.Position[0], model.Camera.Position[1]},
		"Radius":     model.Camera.RadiusToScreen(1.0, constants.ScreenWidth, constants.ScreenHeight),
	}

	opts := &ebiten.DrawRectShaderOptions{
		Uniforms: uniforms,
		GeoM:     geoM,
	}
	opts.Images[0] = offscreen
	screen.DrawRectShader(constants.ScreenWidth, constants.ScreenHeight, r.spiralDistortionShader, opts)

	return true
}

func (r *Renderer) ResetShaderTime() {
	r.startTime = time.Now()
}

func (r *Renderer) applyPortalDistortion(
	screen ebiten.FinalScreen,
	offscreen *ebiten.Image,
	model *Models.SpaceGame,
	geoM ebiten.GeoM,
) {
	if r.portalDistortionShader == nil {
		return
	}

	buf := r.portalBuffer1
	camera := model.Camera
	t := float32(time.Since(r.startTime).Seconds())

	portalColors := []color.RGBA{
		{0, 204, 255, 180},
		{255, 102, 204, 180},
		{204, 255, 51, 180},
		{255, 153, 0, 180},
		{153, 51, 255, 180},
		{255, 255, 51, 180},
	}

	for _, portal := range model.Portals {
		// 1) copy *current* screen contents into buf
		buf.Clear()
		buf.DrawImage(offscreen, nil)

		// 2) build this portal’s uniforms
		idx := portal.PairID % len(portalColors)
		c := portalColors[idx]
		uniforms := map[string]any{
			"Portal_Pos":    []float32{portal.Position[0], portal.Position[1]},
			"Portal_Radius": portal.GetOrbitRadius(),
			"Portal_Color":  []float32{float32(c.R) / 255, float32(c.G) / 255, float32(c.B) / 255},
			"CameraPos":     []float32{camera.Position[0], camera.Position[1]},
			"Zoom":          camera.GetTotalZoom(),
			"Time":          t,
			"ScreenSize":    []float32{float32(constants.ScreenWidth), float32(constants.ScreenHeight)},
			"IsActive": func() float32 {
				if portal.IsActive && portal.CooldownTimer <= 0 {
					return 1
				}
				return 0
			}(),
		}

		// 3) draw full-screen quad, sampling from buf, **onto** screen
		opts := &ebiten.DrawRectShaderOptions{
			Uniforms: uniforms,
			//GeoM:     geoM,
		}
		opts.Images[0] = buf

		screen.DrawRectShader(
			constants.ScreenWidth,
			constants.ScreenHeight,
			r.portalDistortionShader,
			opts,
		)
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
