package rendering

import (
	"bytes"
	_ "embed"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/you/trajectory/constants"
	"github.com/you/trajectory/space/colors"
	Models "github.com/you/trajectory/space/model"
	"github.com/you/trajectory/space/shadows"
	"golang.org/x/image/math/f32"
	"image/color"
	"log"
	"math"
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

//go:embed blackhole.go
var blackHoleShader []byte

//go:embed whitehole.go
var whiteHoleShader []byte

//go:embed spiral_distortion.go
var spiralDistortionShader []byte

//go:embed wormhole.go
var wormholeShader []byte

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

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// minFloat32 returns the minimum of two float32 values
func minFloat32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

// maxFloat32 returns the maximum of two float32 values
func maxFloat32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

const (
	baseArm       = float32(8) // half-length of each arm by default
	stretchFactor = float32(1) // how much longer the interior arms get
	lineWidth     = float32(2.25)
)

// Renderer handles all rendering operations for the game
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
	wormholeShader         *ebiten.Shader
	whiteTexture           *ebiten.Image
	intermediateBuffer     *ebiten.Image // Persistent buffer for post-processing
	startTime              time.Time     // Game start time for shader animations
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

	// Initialize black hole shader
	var blackHoleShaderObj *ebiten.Shader
	if shader, err := ebiten.NewShader(blackHoleShader); err == nil {
		blackHoleShaderObj = shader
	} else {
		log.Println("Failed to load black hole shader:", err)
	}

	// Initialize white hole shader
	var whiteHoleShaderObj *ebiten.Shader
	if shader, err := ebiten.NewShader(whiteHoleShader); err == nil {
		whiteHoleShaderObj = shader
	} else {
		log.Println("Failed to load white hole shader:", err)
	}

	// Initialize spiral distortion shader
	var spiralDistortionShaderObj *ebiten.Shader
	if shader, err := ebiten.NewShader(spiralDistortionShader); err == nil {
		spiralDistortionShaderObj = shader
	} else {
		log.Println("Failed to load spiral distortion shader:", err)
	}

	// Initialize wormhole shader
	var wormholeShaderObj *ebiten.Shader
	if shader, err := ebiten.NewShader(wormholeShader); err == nil {
		wormholeShaderObj = shader
	} else {
		log.Println("Failed to load wormhole shader:", err)
	}

	// Create a white texture matching screen size for shaders that need a source image
	whiteTexture := ebiten.NewImage(constants.ScreenWidth, constants.ScreenHeight)
	whiteTexture.Fill(color.White)

	// Create persistent intermediate buffer for post-processing
	intermediateBuffer := ebiten.NewImage(constants.ScreenWidth, constants.ScreenHeight)

	return &Renderer{
		shadowSystem:           shadows.NewShadowSystem(constants.ScreenWidth, constants.ScreenHeight),
		orbitShader:            orbitShader,
		playerTrailShader:      playerTrailShaderObj,
		trajectoryArrowShader:  trajectoryArrowShaderObj,
		nebulaShader:           nebulaShader,
		invertOnLightShader:    invertOnLightShaderObj,
		revealOnLightShader:    revealOnLightShaderObj,
		alienTrailShader:       alienTrailShaderObj,
		planetShader:           planetShaderObj,
		blackHoleShader:        blackHoleShaderObj,
		whiteHoleShader:        whiteHoleShaderObj,
		spiralDistortionShader: spiralDistortionShaderObj,
		wormholeShader:         wormholeShaderObj,
		whiteTexture:           whiteTexture,
		intermediateBuffer:     intermediateBuffer,
		startTime:              time.Now(), // Initialize start time for shader animations
	}
}

// Draw renders the complete game scene with spiral distortion
func (r *Renderer) Draw(screen *ebiten.Image, model *Models.SpaceGame) {
	// For testing, always apply the shader (remove this condition later)
	if r.spiralDistortionShader != nil {
		// Clear the persistent intermediate buffer
		r.intermediateBuffer.Clear()

		// Render everything to intermediate buffer first (EXCLUDING black holes and portals)
		r.drawNebulaBackground(r.intermediateBuffer, model)
		r.renderShadows(r.intermediateBuffer, model)
		r.drawCelestialBodies(r.intermediateBuffer, model) // Now excludes black holes
		r.drawAsteroids(r.intermediateBuffer, model)
		r.drawPortals(r.intermediateBuffer, model) // Simple portals for intermediate buffer
		r.drawPlayerTrail(r.intermediateBuffer, model)
		r.drawPlayer(r.intermediateBuffer, model)
		r.drawTrajectoryArrow(r.intermediateBuffer, model)
		r.drawLastCollisionMarker(r.intermediateBuffer, model)
		r.drawBorderIndicators(r.intermediateBuffer, model)
		r.drawFPS(r.intermediateBuffer)

		// Apply post-processing effect to final screen
		blackHoles := r.getBlackHoles(model)
		r.applySpiralOverlay(screen, r.intermediateBuffer, model, blackHoles)

		// Draw black holes AFTER the spiral distortion effect
		r.DrawBlackHoles(screen, model)
		
		// Draw wormhole portals AFTER everything else to get the final effect
		r.drawPortals(screen, model)
	} else {
		// No shader available, render directly
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
		// When no shader, still draw black holes normally
		r.DrawBlackHoles(screen, model)
	}
}

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

// drawPortals renders all portals in the game using wormhole effect
func (r *Renderer) drawPortals(screen *ebiten.Image, model *Models.SpaceGame) {
	// Check if we're drawing to the intermediate buffer (during the first pass)
	// If so, use simple portal rendering to avoid circular reference
	if screen == r.intermediateBuffer {
		r.drawPortalsSimple(screen, model)
		return
	}

	if r.wormholeShader == nil {
		// Fallback to simple circles if shader is not available
		r.drawPortalsSimple(screen, model)
		return
	}

	// Group portals by pairs
	portalPairs := make(map[int][]*Models.Portal)
	for _, portal := range model.Portals {
		portalPairs[portal.PairID] = append(portalPairs[portal.PairID], portal)
	}

	// Draw each portal pair with wormhole effect
	for _, pair := range portalPairs {
		if len(pair) == 2 {
			r.drawPortalPair(screen, pair[0], pair[1], model)
		}
	}
}

// drawPortalsSimple renders portals as simple circles (fallback)
func (r *Renderer) drawPortalsSimple(screen *ebiten.Image, model *Models.SpaceGame) {
	currentTime := float32(time.Since(r.startTime).Seconds())
	for _, portal := range model.Portals {
		r.drawPortalSimple(screen, portal, model, currentTime)
	}
}

// drawPortalDebug draws a simple wireframe outline of the portal for debugging
func (r *Renderer) drawPortalDebug(screen *ebiten.Image, portal *Models.Portal, model *Models.SpaceGame) {
	camera := model.Camera

	// Convert portal position to screen coordinates
	screenPos := camera.WorldToScreen(portal.Position, constants.ScreenWidth, constants.ScreenHeight)

	// Convert portal dimensions to screen pixels
	halfWidth := camera.RadiusToScreen(portal.Width/2, constants.ScreenWidth, constants.ScreenHeight)
	halfHeight := camera.RadiusToScreen(portal.Height/2, constants.ScreenWidth, constants.ScreenHeight)

	// Portal colors for debug outline
	debugColors := []color.RGBA{
		{R: 0, G: 255, B: 255, A: 128},   // Cyan
		{R: 255, G: 100, B: 200, A: 128}, // Pink
		{R: 200, G: 255, B: 50, A: 128},  // Lime
		{R: 255, G: 150, B: 0, A: 128},   // Orange
		{R: 150, G: 50, B: 255, A: 128},  // Purple
		{R: 255, G: 255, B: 50, A: 128},  // Yellow
	}

	colorIndex := portal.PairID % len(debugColors)
	debugColor := debugColors[colorIndex]

	// Dim color if on cooldown
	if portal.CooldownTimer > 0 {
		debugColor.A = 64
	}

	// Draw simple rectangle outline
	r.drawRectOutline(screen, screenPos[0]-halfWidth, screenPos[1]-halfHeight, halfWidth*2, halfHeight*2, 2.0, debugColor)
}

// drawRectOutline draws a rectangle outline
func (r *Renderer) drawRectOutline(screen *ebiten.Image, x, y, width, height, lineWidth float32, color color.RGBA) {
	// Top edge
	r.drawLine(screen, x, y, x+width, y, lineWidth, color)
	// Bottom edge
	r.drawLine(screen, x, y+height, x+width, y+height, lineWidth, color)
	// Left edge
	r.drawLine(screen, x, y, x, y+height, lineWidth, color)
	// Right edge
	r.drawLine(screen, x+width, y, x+width, y+height, lineWidth, color)
}

// drawPortalSimple renders a single portal as a simple circle (fallback method)
func (r *Renderer) drawPortalSimple(screen *ebiten.Image, portal *Models.Portal, model *Models.SpaceGame, currentTime float32) {
	// Convert portal position to screen coordinates using camera methods
	camera := model.Camera
	worldPos := portal.Position
	screenPos := camera.WorldToScreen(worldPos, constants.ScreenWidth, constants.ScreenHeight)

	// Use the average of width and height as the circle radius
	portalRadius := camera.RadiusToScreen((portal.Width+portal.Height)/4, constants.ScreenWidth, constants.ScreenHeight)

	// Portal colors (different colors for different pair IDs)
	portalColors := []color.RGBA{
		{R: 0, G: 204, B: 255, A: 180},   // Cyan
		{R: 255, G: 102, B: 204, A: 180}, // Pink
		{R: 204, G: 255, B: 51, A: 180},  // Lime
		{R: 255, G: 153, B: 0, A: 180},   // Orange
		{R: 153, G: 51, B: 255, A: 180},  // Purple
		{R: 255, G: 255, B: 51, A: 180},  // Yellow
	}

	// Select color based on pair ID
	colorIndex := portal.PairID % len(portalColors)
	portalColor := portalColors[colorIndex]

	// Adjust alpha based on activity state
	if !portal.IsActive || portal.CooldownTimer > 0 {
		portalColor.A = 60 // Dim when inactive or on cooldown
	} else if portal.PlayerInside {
		portalColor.A = 120 // Semi-dim when player is inside (won't teleport until they exit and re-enter)
	}

	// Add pulsing effect
	pulseIntensity := 0.7 + 0.3*float32(math.Sin(float64(currentTime*3.0)))
	finalColor := color.RGBA{
		R: uint8(float32(portalColor.R) * pulseIntensity),
		G: uint8(float32(portalColor.G) * pulseIntensity),
		B: uint8(float32(portalColor.B) * pulseIntensity),
		A: portalColor.A,
	}

	// Draw outer glow circle (larger, more transparent)
	glowRadius := portalRadius * 4
	glowColor := color.RGBA{R: finalColor.R, G: finalColor.G, B: finalColor.B, A: finalColor.A / 4}
	vector.DrawFilledCircle(screen, screenPos[0], screenPos[1], glowRadius, glowColor, true)

	// Draw main portal circle
	vector.DrawFilledCircle(screen, screenPos[0], screenPos[1], portalRadius, finalColor, true)

	// Draw inner bright core
	coreRadius := portalRadius * 0.3
	coreColor := color.RGBA{R: 255, G: 255, B: 255, A: uint8(float32(finalColor.A) * 0.8)}
	vector.DrawFilledCircle(screen, screenPos[0], screenPos[1], coreRadius, coreColor, true)
}

// drawPortalPair renders a pair of portals with wormhole effect using shader
func (r *Renderer) drawPortalPair(screen *ebiten.Image, portalA, portalB *Models.Portal, model *Models.SpaceGame) {
	camera := model.Camera
	
	// Portal colors (different colors for different pair IDs)
	portalColors := []f32.Vec3{
		{0.0, 0.8, 1.0}, // Cyan
		{1.0, 0.4, 0.8}, // Pink
		{0.8, 1.0, 0.2}, // Lime
		{1.0, 0.6, 0.0}, // Orange
		{0.6, 0.2, 1.0}, // Purple
		{1.0, 1.0, 0.2}, // Yellow
	}
	
	// Select color based on pair ID
	colorIndex := portalA.PairID % len(portalColors)
	portalColor := portalColors[colorIndex]
	
	// Convert portal positions to screen coordinates
	screenPosA := camera.WorldToScreen(portalA.Position, constants.ScreenWidth, constants.ScreenHeight)
	screenPosB := camera.WorldToScreen(portalB.Position, constants.ScreenWidth, constants.ScreenHeight)
	
	// Calculate portal radii in screen space
	radiusA := camera.RadiusToScreen((portalA.Width+portalA.Height)/4, constants.ScreenWidth, constants.ScreenHeight)
	radiusB := camera.RadiusToScreen((portalB.Width+portalB.Height)/4, constants.ScreenWidth, constants.ScreenHeight)
	
	// Determine render area that encompasses both portals with some padding
	padding := float32(100)
	minX := minFloat32(screenPosA[0]-radiusA-padding, screenPosB[0]-radiusB-padding)
	maxX := maxFloat32(screenPosA[0]+radiusA+padding, screenPosB[0]+radiusB+padding)
	minY := minFloat32(screenPosA[1]-radiusA-padding, screenPosB[1]+radiusB+padding)
	maxY := maxFloat32(screenPosA[1]+radiusA+padding, screenPosB[1]+radiusB+padding)
	
	// Clamp to screen bounds
	minX = maxFloat32(minX, 0)
	maxX = minFloat32(maxX, float32(constants.ScreenWidth))
	minY = maxFloat32(minY, 0)
	maxY = minFloat32(maxY, float32(constants.ScreenHeight))
	
	renderWidth := int(maxX - minX)
	renderHeight := int(maxY - minY)
	
	if renderWidth <= 0 || renderHeight <= 0 {
		return // Nothing to render
	}
	
	// Set up shader options
	opts := &ebiten.DrawTrianglesShaderOptions{}
	opts.Uniforms = map[string]interface{}{
		"PortalA_Pos":      []float32{portalA.Position[0], portalA.Position[1]},
		"PortalA_Radius":   (portalA.Width + portalA.Height) / 4.0,
		"PortalA_Rotation": portalA.Rotation,
		"PortalA_Color":    []float32{portalColor[0], portalColor[1], portalColor[2]},
		
		"PortalB_Pos":      []float32{portalB.Position[0], portalB.Position[1]},
		"PortalB_Radius":   (portalB.Width + portalB.Height) / 4.0,
		"PortalB_Rotation": portalB.Rotation,
		"PortalB_Color":    []float32{portalColor[0], portalColor[1], portalColor[2]},
		
		"CameraPos":   []float32{camera.Position[0], camera.Position[1]},
		"Zoom":        camera.GetEffectiveZoom(),
		"Time":        float32(time.Since(r.startTime).Seconds()),
		"ScreenSize":  []float32{constants.ScreenWidth, constants.ScreenHeight},
		"IsActive":    func() float32 { if portalA.IsActive && portalB.IsActive { return 1.0 } else { return 0.0 } }(),
	}
	
	// Set blending mode for portal effect
	opts.Blend = ebiten.BlendSourceOver
	
	// Create vertices for the render area
	vertices := []ebiten.Vertex{
		{DstX: minX, DstY: minY, SrcX: minX, SrcY: minY, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: maxX, DstY: minY, SrcX: maxX, SrcY: minY, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: minX, DstY: maxY, SrcX: minX, SrcY: maxY, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: maxX, DstY: maxY, SrcX: maxX, SrcY: maxY, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
	}
	indices := []uint16{0, 1, 2, 1, 2, 3}
	
	// Apply wormhole shader with the intermediate buffer as source
	opts.Images[0] = r.intermediateBuffer
	screen.DrawTrianglesShader(vertices, indices, r.wormholeShader, opts)
}
