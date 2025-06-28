package rendering

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/you/trajectory/constants"
	"github.com/you/trajectory/space/colors"
	Models "github.com/you/trajectory/space/model"
	"github.com/you/trajectory/space/resources"
	"github.com/you/trajectory/space/shadows"
	"github.com/you/trajectory/space/util"
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

// Renderer handles all rendering operations for the game
type Renderer struct {
	shadowSystem          *shadows.ShadowSystem
	orbitShader           *ebiten.Shader
	playerTrailShader     *ebiten.Shader
	trajectoryArrowShader *ebiten.Shader
	nebulaShader          *ebiten.Shader
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

	// Create a white texture matching screen size for shaders that need a source image
	whiteTexture := ebiten.NewImage(constants.ScreenWidth, constants.ScreenHeight)
	whiteTexture.Fill(color.White)

	return &Renderer{
		shadowSystem:          shadows.NewShadowSystem(constants.ScreenWidth, constants.ScreenHeight),
		orbitShader:           orbitShader,
		playerTrailShader:     playerTrailShaderObj,
		trajectoryArrowShader: trajectoryArrowShaderObj,
		nebulaShader:          nebulaShader,
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

	r.shadowSystem.RenderShadows(screen, lightPos, lightDirection, fov, camera.Zoom, celestialPositions, celestialRadii, asteroidPositions, asteroidRadii, false)
}

// drawCelestialBodies renders all celestial bodies
func (r *Renderer) drawCelestialBodies(screen *ebiten.Image, model *Models.SpaceGame) {
	camera := model.Camera

	for _, body := range model.CelestialBodies {
		bodyPos := body.GetPosition()
		screenPos := camera.WorldToScreen(bodyPos, constants.ScreenWidth, constants.ScreenHeight)
		radius := camera.RadiusToScreen(body.GetRadius(), constants.ScreenWidth, constants.ScreenHeight)

		// Choose colors based on celestial body type
		var bodyColor color.RGBA
		var orbitColor color.RGBA

		switch body.GetType() {
		case Models.CelestialBodyTypePlanet:
			bodyColor = colors.PlanetBody
			orbitColor = colors.PlanetOrbit
		case Models.CelestialBodyTypeBlackHole:
			bodyColor = colors.BlackHoleBody
			orbitColor = colors.BlackHoleOrbit
		case Models.CelestialBodyTypeWhiteHole:
			bodyColor = colors.WhiteHoleBody
			orbitColor = colors.WhiteHoleOrbit
		case Models.CelestialBodyTypeAsteroid:
			bodyColor = colors.AsteroidBodyAlt
			orbitColor = colors.AsteroidOrbit
		}

		// Check if this celestial body has an image
		var imagePath string
		switch body.GetType() {
		case Models.CelestialBodyTypePlanet:
			if planet, ok := body.(*Models.Planet); ok && planet.ImagePath != "" {
				imagePath = planet.ImagePath
			}
		case Models.CelestialBodyTypeBlackHole:
			if blackHole, ok := body.(*Models.BlackHole); ok && blackHole.ImagePath != "" {
				imagePath = blackHole.ImagePath
			}
		case Models.CelestialBodyTypeWhiteHole:
			if whiteHole, ok := body.(*Models.WhiteHole); ok && whiteHole.ImagePath != "" {
				imagePath = whiteHole.ImagePath
			}
		}

		if imagePath != "" {
			// Render with image
			r.drawCelestialBodyWithImage(screen, screenPos, radius, imagePath)
		} else {
			// Fallback to circle rendering
			vector.DrawFilledCircle(screen, screenPos[0], screenPos[1], radius, bodyColor, true)
		}

		// Draw celestial body's orbit radius as a dashed circle with light effect
		r.drawOrbitCircleWithLight(screen, model, screenPos, camera.RadiusToScreen(body.GetOrbitRadius(), constants.ScreenWidth, constants.ScreenHeight), orbitColor)
	}
}

// drawAsteroids renders all asteroids
func (r *Renderer) drawAsteroids(screen *ebiten.Image, model *Models.SpaceGame) {
	camera := model.Camera

	for _, asteroid := range model.RingAsteroids {
		asteroidPos := asteroid.GetPosition()
		screenPos := camera.WorldToScreen(asteroidPos, constants.ScreenWidth, constants.ScreenHeight)
		radius := camera.RadiusToScreen(asteroid.GetRadius(), constants.ScreenWidth, constants.ScreenHeight)

		if asteroid.ImagePath != "" {
			// Render with image
			r.drawCelestialBodyWithImage(screen, screenPos, radius, asteroid.ImagePath)
		} else {
			// Fallback to circle rendering
			vector.DrawFilledCircle(screen, screenPos[0], screenPos[1], radius, colors.AsteroidBody, true)
		}
	}
}

// drawPlayer renders the player
func (r *Renderer) drawPlayer(screen *ebiten.Image, model *Models.SpaceGame) {
	camera := model.Camera
	player := model.Player
	playerScreenPos := camera.WorldToScreen(player.Position, constants.ScreenWidth, constants.ScreenHeight)
	playerRadius := camera.RadiusToScreen(player.Radius, constants.ScreenWidth, constants.ScreenHeight)

	// Render player with image if ImagePath is provided, otherwise use green circle
	if player.ImagePath != "" {
		playerImg := resources.LoadImage(player.ImagePath)
		if playerImg != nil {
			// Calculate rotation based on velocity direction
			var rotation float64 = 0
			if player.IsMoving() {
				rotation = math.Atan2(float64(player.Velocity[1]), float64(player.Velocity[0]))
			}
			r.drawPlayerWithImage(screen, playerScreenPos, playerRadius, player.ImagePath, rotation)
		} else {
			// Fallback to green circle if image loading fails
			vector.DrawFilledCircle(screen, playerScreenPos[0], playerScreenPos[1], playerRadius, colors.PlayerBody, true)
		}
	} else {
		// Draw simple green circle when no ImagePath is provided
		vector.DrawFilledCircle(screen, playerScreenPos[0], playerScreenPos[1], playerRadius, colors.PlayerBody, true)
	}
}

// drawPlayerTrail draws the player's movement trail with fading effect
func (r *Renderer) drawPlayerTrail(screen *ebiten.Image, model *Models.SpaceGame) {
	trailPoints := model.Player.GetTrailPoints()
	if len(trailPoints) < 2 {
		return // Need at least 2 points to draw lines
	}

	camera := model.Camera
	now := time.Now()

	// If shadows are enabled and we have the player trail shader, apply light effects
	if model.ShadowsEnabled && r.playerTrailShader != nil {
		// Get light information (same as shadow system)
		lightPos := camera.WorldToScreen(model.Player.Position, constants.ScreenWidth, constants.ScreenHeight)
		lightDirection := camera.WorldToScreen(camera.Position, constants.ScreenWidth, constants.ScreenHeight)

		// Calculate light direction vector (from light pos to camera/target)
		lightDirVec := f32.Vec2{
			lightDirection[0] - lightPos[0],
			lightDirection[1] - lightPos[1],
		}

		// Calculate max distance for the light cone (same as shadow system)
		maxDistance := math.Hypot(float64(constants.ScreenWidth), float64(constants.ScreenHeight))
		fov := r.getAdaptiveFov(lightDirection, lightPos)

		// Draw lines between consecutive trail points with shader
		for i := 1; i < len(trailPoints); i++ {
			prevPoint := trailPoints[i-1]
			currPoint := trailPoints[i]

			// Calculate age of the current point
			age := now.Sub(currPoint.Timestamp)
			trailDuration := 3.0 * time.Second // Should match the constant in Player.go

			// Calculate alpha based on age (newer = more opaque)
			ageRatio := float64(age) / float64(trailDuration)
			alpha := uint8(255 * (1.0 - ageRatio))
			if ageRatio >= 1.0 {
				continue // Skip expired points
			}

			// Convert world positions to screen coordinates
			prevScreen := camera.WorldToScreen(prevPoint.Position, constants.ScreenWidth, constants.ScreenHeight)
			currScreen := camera.WorldToScreen(currPoint.Position, constants.ScreenWidth, constants.ScreenHeight)

			// Trail color with fading alpha
			trailColor := color.RGBA{R: colors.PlayerTrail.R, G: colors.PlayerTrail.G, B: colors.PlayerTrail.B, A: alpha}

			// Calculate line width based on age (newer = thicker)
			width := float32(1 + (1.0-ageRatio)*2) // Width from 1 to 3

			// Prepare shader uniforms
			uniforms := map[string]any{
				"LightPos":       []float32{lightPos[0], lightPos[1]},
				"LightDirection": []float32{lightDirVec[0], lightDirVec[1]},
				"FOVAngle":       float32(fov * math.Pi / 180.0), // Convert to radians
				"MaxDistance":    float32(maxDistance),
				"Zoom":           camera.Zoom,
				"OriginalColor": []float32{
					float32(trailColor.R) / 255.0,
					float32(trailColor.G) / 255.0,
					float32(trailColor.B) / 255.0,
					float32(trailColor.A) / 255.0,
				},
			}

			// Draw the trail segment with shader effect
			r.drawLineWithShader(screen, prevScreen, currScreen, width, trailColor, r.playerTrailShader, uniforms)
		}
	}
}

// drawTrajectoryArrow draws the trajectory arrow when dragging
func (r *Renderer) drawTrajectoryArrow(screen *ebiten.Image, model *Models.SpaceGame) {
	// This would need access to input dragInfo - will be handled in the main game loop
	// For now, this is a placeholder for the trajectory arrow rendering logic
}

// drawOrbitCircleWithLight draws a dashed orbit circle with light inversion effect
func (r *Renderer) drawOrbitCircleWithLight(screen *ebiten.Image, model *Models.SpaceGame, center f32.Vec2, radius float32, orbitColor color.RGBA) {
	if radius <= 0 {
		return
	}

	// Dashed circle parameters
	const numDashes = 24
	const dashPortion = 4.0 / (4.0 + 16.0) // = 0.2

	// Calculate dash and gap lengths
	circ := 2 * math.Pi * float64(radius)
	segLen := float32(circ) / float32(numDashes)
	dashLen := segLen * dashPortion
	gapLen := segLen * (1.0 - dashPortion)

	// If shadows are enabled and we have the orbit shader, apply light effects
	if model.ShadowsEnabled && r.orbitShader != nil {
		camera := model.Camera

		// Get light information (same as shadow system)
		lightPos := camera.WorldToScreen(model.Player.Position, constants.ScreenWidth, constants.ScreenHeight)
		lightDirection := camera.WorldToScreen(camera.Position, constants.ScreenWidth, constants.ScreenHeight)

		// Calculate light direction vector (from light pos to camera/target)
		lightDirVec := f32.Vec2{
			lightDirection[0] - lightPos[0],
			lightDirection[1] - lightPos[1],
		}

		// Calculate max distance for the light cone (same as shadow system)
		maxDistance := math.Hypot(float64(constants.ScreenWidth), float64(constants.ScreenHeight))

		// Calculate elapsed time since game start for rotation animation
		currentTime := float32(time.Since(r.startTime).Seconds())

		fov := r.getAdaptiveFov(lightDirection, lightPos)

		// Prepare shader uniforms
		uniforms := map[string]any{
			"LightPos":       []float32{lightPos[0], lightPos[1]},
			"LightDirection": []float32{lightDirVec[0], lightDirVec[1]},
			"FOVAngle":       float32(fov * math.Pi / 180.0), // Convert to radians
			"MaxDistance":    float32(maxDistance),
			"Zoom":           camera.Zoom,
			"OriginalColor": []float32{
				float32(orbitColor.R) / 255.0,
				float32(orbitColor.G) / 255.0,
				float32(orbitColor.B) / 255.0,
				float32(orbitColor.A) / 255.0,
			},
			"Time":              currentTime,
			"RotationDirection": float32(1.0), // Counterclockwise rotation
			"CircleCenter":      []float32{center[0], center[1]},
			"CircleRadius":      radius,
		}

		// Use shader-enabled dashed circle
		util.StrokeDashedCircleTrianglesWithShader(screen, center[0], center[1], radius, 4, orbitColor, dashLen, gapLen, true, r.orbitShader, uniforms, currentTime/10.0)
	} else {
		// Fallback to regular dashed circle
		util.StrokeDashedCircle(screen, center[0], center[1], radius, 4, orbitColor, dashLen, gapLen, true)
	}
}

// drawOrbitCircle draws a dashed orbit circle
func (r *Renderer) drawOrbitCircle(screen *ebiten.Image, center f32.Vec2, radius float32, color color.RGBA) {
	if radius <= 0 {
		return
	}

	// Dashed circle parameters
	const numDashes = 24
	const dashPortion = 4.0 / (4.0 + 16.0) // = 0.2

	// Calculate dash and gap lengths
	circ := 2 * math.Pi * float64(radius)
	segLen := float32(circ) / float32(numDashes)
	dashLen := segLen * dashPortion
	gapLen := segLen * (1.0 - dashPortion)

	util.StrokeDashedCircle(screen, center[0], center[1], radius, 4, color, dashLen, gapLen, true)
}

// drawArrowHead draws an arrowhead at the end of a line
func (r *Renderer) drawArrowHead(screen *ebiten.Image, start, end f32.Vec2, color color.RGBA) {
	// Calculate arrow direction
	dx := end[0] - start[0]
	dy := end[1] - start[1]
	length := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	if length == 0 {
		return
	}

	// Normalize direction vector
	unitX := dx / length
	unitY := dy / length

	// Arrow head size
	arrowLength := float32(15.0)
	arrowWidth := float32(8.0)

	// Calculate perpendicular vector
	perpX := -unitY
	perpY := unitX

	// Calculate arrow head points
	backX := end[0] - unitX*arrowLength
	backY := end[1] - unitY*arrowLength

	leftX := backX + perpX*arrowWidth
	leftY := backY + perpY*arrowWidth

	rightX := backX - perpX*arrowWidth
	rightY := backY - perpY*arrowWidth

	// Draw arrow head lines
	vector.StrokeLine(screen, end[0], end[1], leftX, leftY, 3, color, true)
	vector.StrokeLine(screen, end[0], end[1], rightX, rightY, 3, color, true)
}

// drawCelestialBodyWithImage renders a celestial body using an image texture
func (r *Renderer) drawCelestialBodyWithImage(screen *ebiten.Image, screenPos f32.Vec2, radius float32, imagePath string) {
	// Load the image
	img := resources.LoadImage(imagePath)
	if img == nil {
		// Fallback to circle if image loading fails
		vector.DrawFilledCircle(screen, screenPos[0], screenPos[1], radius, colors.PlanetBody, true)
		return
	}

	// Calculate scaling to fit the desired radius
	imgSize := img.Bounds().Size()
	imgRadius := float32(imgSize.X) / 2.0 // Assume square images
	if imgSize.Y > imgSize.X {
		imgRadius = float32(imgSize.Y) / 2.0
	}

	scale := (radius * 2.0) / (imgRadius * 2.0) // Scale to fit diameter

	// Create draw options
	op := &ebiten.DrawImageOptions{}

	// Move image center to origin for rotation/scaling
	op.GeoM.Translate(-float64(imgSize.X)/2, -float64(imgSize.Y)/2)

	// Scale the image to the desired size
	op.GeoM.Scale(float64(scale), float64(scale))

	// Move to final screen position
	op.GeoM.Translate(float64(screenPos[0]), float64(screenPos[1]))

	// Draw the image
	screen.DrawImage(img, op)
}

// drawPlayerWithImage renders the player using an image texture with rotation based on movement
func (r *Renderer) drawPlayerWithImage(screen *ebiten.Image, screenPos f32.Vec2, radius float32, imagePath string, rotation float64) {
	// Load the image
	img := resources.LoadImage(imagePath)
	if img == nil {
		// Fallback to circle if image loading fails
		vector.DrawFilledCircle(screen, screenPos[0], screenPos[1], radius, colors.PlayerBody, true)
		return
	}

	// Calculate scaling to fit the desired radius
	imgSize := img.Bounds().Size()
	imgRadius := float32(imgSize.X) / 2.0 // Assume square images
	if imgSize.Y > imgSize.X {
		imgRadius = float32(imgSize.Y) / 2.0
	}

	scale := (radius * 2.0) / (imgRadius * 2.0) // Scale to fit diameter

	// Create draw options
	op := &ebiten.DrawImageOptions{}

	// Move image center to origin for rotation/scaling
	op.GeoM.Translate(-float64(imgSize.X)/2, -float64(imgSize.Y)/2)

	// Apply rotation (astronaut points in direction of movement)
	op.GeoM.Rotate(rotation)

	// Scale the image to the desired size
	op.GeoM.Scale(float64(scale), float64(scale))

	// Move to final screen position
	op.GeoM.Translate(float64(screenPos[0]), float64(screenPos[1]))

	// Draw the image
	screen.DrawImage(img, op)
}

// DrawTrajectoryArrow draws trajectory arrow (called from main game loop with drag info)
func (r *Renderer) DrawTrajectoryArrow(screen *ebiten.Image, model *Models.SpaceGame, startPos, currentPos f32.Vec2, isDragging bool) {
	if !isDragging || model.Player.IsMoving() {
		return
	}

	camera := model.Camera

	// Convert screen drag to world space for trajectory calculation
	startWorld := camera.ScreenToWorld(startPos, constants.ScreenWidth, constants.ScreenHeight)
	currentWorld := camera.ScreenToWorld(currentPos, constants.ScreenWidth, constants.ScreenHeight)

	// Calculate trajectory vector (opposite of drag direction)
	trajectoryVector := f32.Vec2{
		startWorld[0] - currentWorld[0],
		startWorld[1] - currentWorld[1],
	}

	// Calculate drag distance and clamp it to maximum allowed (same as throw logic)
	const maxDragDistance = float32(0.25)
	const maxVelocity = float32(0.5)

	trajectoryDistance := float32(math.Sqrt(float64(trajectoryVector[0]*trajectoryVector[0] + trajectoryVector[1]*trajectoryVector[1])))
	if trajectoryDistance > maxDragDistance {
		// Normalize and scale to max distance
		trajectoryVector[0] = (trajectoryVector[0] / trajectoryDistance) * maxDragDistance
		trajectoryVector[1] = (trajectoryVector[1] / trajectoryDistance) * maxDragDistance
	}

	// Scale trajectory for visualization (same as throw logic)
	velocityMultiplier := float32(2.0)
	trajectoryVector[0] *= velocityMultiplier
	trajectoryVector[1] *= velocityMultiplier

	// Clamp final visualization vector to match max velocity
	visualMagnitude := float32(math.Sqrt(float64(trajectoryVector[0]*trajectoryVector[0] + trajectoryVector[1]*trajectoryVector[1])))
	if visualMagnitude > maxVelocity {
		trajectoryVector[0] = (trajectoryVector[0] / visualMagnitude) * maxVelocity
		trajectoryVector[1] = (trajectoryVector[1] / visualMagnitude) * maxVelocity
	}

	// Start arrow from player center
	playerWorldPos := model.Player.Position
	endWorld := f32.Vec2{
		playerWorldPos[0] + trajectoryVector[0],
		playerWorldPos[1] + trajectoryVector[1],
	}

	// Convert to screen coordinates
	playerScreen := camera.WorldToScreen(playerWorldPos, constants.ScreenWidth, constants.ScreenHeight)
	endScreen := camera.WorldToScreen(endWorld, constants.ScreenWidth, constants.ScreenHeight)

	// Only draw if trajectory has significant length
	dragDistance := float32(math.Sqrt(float64(trajectoryVector[0]*trajectoryVector[0] + trajectoryVector[1]*trajectoryVector[1])))
	if dragDistance > 0.01 {
		// If shadows are enabled and we have the trajectory arrow shader, apply light effects
		if model.ShadowsEnabled && r.trajectoryArrowShader != nil {
			// Get light information (same as shadow system)
			lightPos := camera.WorldToScreen(model.Player.Position, constants.ScreenWidth, constants.ScreenHeight)
			lightDirection := camera.WorldToScreen(camera.Position, constants.ScreenWidth, constants.ScreenHeight)

			// Calculate light direction vector (from light pos to camera/target)
			lightDirVec := f32.Vec2{
				lightDirection[0] - lightPos[0],
				lightDirection[1] - lightPos[1],
			}

			// Calculate max distance for the light cone (same as shadow system)
			maxDistance := math.Hypot(float64(constants.ScreenWidth), float64(constants.ScreenHeight))

			fov := r.getAdaptiveFov(lightDirection, lightPos)

			// Prepare shader uniforms
			uniforms := map[string]any{
				"LightPos":       []float32{lightPos[0], lightPos[1]},
				"LightDirection": []float32{lightDirVec[0], lightDirVec[1]},
				"FOVAngle":       float32(fov * math.Pi / 180.0), // Convert to radians
				"MaxDistance":    float32(maxDistance),
				"Zoom":           camera.Zoom,
				"OriginalColor": []float32{
					float32(colors.TrajectoryArrow.R) / 255.0,
					float32(colors.TrajectoryArrow.G) / 255.0,
					float32(colors.TrajectoryArrow.B) / 255.0,
					float32(colors.TrajectoryArrow.A) / 255.0,
				},
			}

			// Draw main arrow line with shader effect
			r.drawLineWithShader(screen, playerScreen, endScreen, 4, colors.TrajectoryArrow, r.trajectoryArrowShader, uniforms)

			// Draw arrowhead with shader effect
			r.drawArrowHeadWithShader(screen, playerScreen, endScreen, colors.TrajectoryArrow, r.trajectoryArrowShader, uniforms)
		} else {
			// Fallback to regular trajectory arrow rendering
			// Draw main arrow line (white, thick)
			vector.StrokeLine(screen, playerScreen[0], playerScreen[1], endScreen[0], endScreen[1], 4, colors.TrajectoryArrow, true)

			// Draw arrowhead
			r.drawArrowHead(screen, playerScreen, endScreen, colors.TrajectoryArrow)
		}
	}
}

func (r *Renderer) getAdaptiveFov(lightDirection f32.Vec2, lightPos f32.Vec2) float64 {
	lightDistanceFromCamera := math.Hypot(float64(lightDirection[0]-lightPos[0]), float64(lightDirection[1]-lightPos[1]))
	// Clamp FOV to reasonable range
	return util.Clamp(fovLight/(lightDistanceFromCamera/250.0), 10.0, 90.0)
}

// drawFPS draws the current FPS in the top right corner
func (r *Renderer) drawFPS(screen *ebiten.Image) {
	const fontSize = 16

	// Calculate FPS
	fps := ebiten.ActualTPS()
	fpsText := fmt.Sprintf("FPS: %.0f", fps)

	// Create text face
	face := &text.GoTextFace{
		Source: mplusFaceSource,
		Size:   fontSize,
	}

	// Measure text width to position it correctly in top right
	textWidth, _ := text.Measure(fpsText, face, 0)

	// Position in top right corner with some padding
	const padding = 10
	x := float64(constants.ScreenWidth) - textWidth - padding
	y := float64(padding + fontSize)

	// Draw the FPS text
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, fpsText, face, op)
}

// drawNebulaBackground renders the animated nebula background
func (r *Renderer) drawNebulaBackground(screen *ebiten.Image, model *Models.SpaceGame) {
	if r.nebulaShader == nil {
		// Fallback to solid background
		screen.Fill(colors.NebulaBackground)
		return
	}

	camera := model.Camera

	// Calculate elapsed time since game start for animations
	currentTime := float32(time.Since(r.startTime).Seconds())

	// Get player and camera positions for parallax
	playerPos := model.Player.Position
	cameraPos := camera.Position

	// Prepare shader uniforms
	uniforms := map[string]any{
		"Time":       currentTime,
		"PlayerPos":  []float32{playerPos[0], playerPos[1]}, // World coordinates
		"CameraPos":  []float32{cameraPos[0], cameraPos[1]}, // World coordinates
		"ScreenSize": []float32{float32(constants.ScreenWidth), float32(constants.ScreenHeight)},
		"Zoom":       camera.Zoom,
	}

	// Draw the nebula background shader
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = uniforms
	op.Images[0] = r.whiteTexture // Use white texture as dummy source

	screen.DrawRectShader(constants.ScreenWidth, constants.ScreenHeight, r.nebulaShader, op)
}

// drawLineWithShader draws a line using triangles with the specified shader
func (r *Renderer) drawLineWithShader(screen *ebiten.Image, start, end f32.Vec2, width float32, color color.RGBA, shader *ebiten.Shader, uniforms map[string]any) {
	// Create line using vector path and triangulation
	var path vector.Path
	path.MoveTo(start[0], start[1])
	path.LineTo(end[0], end[1])

	strokeOp := &vector.StrokeOptions{Width: width}

	// Create vertices and indices for the line
	var vs []ebiten.Vertex
	var is []uint16
	vs, is = path.AppendVerticesAndIndicesForStroke(vs, is, strokeOp)

	// Set color for all vertices
	rr, gg, bb, aa := color.RGBA()
	cr := float32(rr) / 0xffff
	cg := float32(gg) / 0xffff
	cb := float32(bb) / 0xffff
	ca := float32(aa) / 0xffff
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = cr
		vs[i].ColorG = cg
		vs[i].ColorB = cb
		vs[i].ColorA = ca
	}

	// Draw with shader
	op := &ebiten.DrawTrianglesShaderOptions{}
	op.Uniforms = uniforms
	op.Images[0] = r.whiteTexture
	screen.DrawTrianglesShader(vs, is, shader, op)
}

// drawBorderIndicators draws "+" markers at the four border corners
func (r *Renderer) drawBorderIndicators(screen *ebiten.Image, model *Models.SpaceGame) {
	// Get border positions from the game model
	border := model.CalculateBorders()
	camera := model.Camera

	// Convert world coordinates to screen coordinates
	bottomLeftScreen := camera.WorldToScreen(border.BottomLeft, constants.ScreenWidth, constants.ScreenHeight)
	bottomRightScreen := camera.WorldToScreen(border.BottomRight, constants.ScreenWidth, constants.ScreenHeight)
	topLeftScreen := camera.WorldToScreen(border.TopLeft, constants.ScreenWidth, constants.ScreenHeight)
	topRightScreen := camera.WorldToScreen(border.TopRight, constants.ScreenWidth, constants.ScreenHeight)

	// Create semi-transparent white color for the border markers
	borderColor := colors.BorderIndicator

	// Size of the "+" markers
	const markerSize = float32(15.0)
	const lineWidth = float32(2.0)

	// Draw "+" at each corner
	r.drawPlusMarker(screen, bottomLeftScreen, markerSize, lineWidth, borderColor)
	r.drawPlusMarker(screen, bottomRightScreen, markerSize, lineWidth, borderColor)
	r.drawPlusMarker(screen, topLeftScreen, markerSize, lineWidth, borderColor)
	r.drawPlusMarker(screen, topRightScreen, markerSize, lineWidth, borderColor)
}

// drawPlusMarker draws a "+" marker at the specified position
func (r *Renderer) drawPlusMarker(screen *ebiten.Image, position f32.Vec2, size, lineWidth float32, color color.RGBA) {
	halfSize := size / 2.0

	// Draw horizontal line
	vector.StrokeLine(screen,
		position[0]-halfSize, position[1],
		position[0]+halfSize, position[1],
		lineWidth, color, true)

	// Draw vertical line
	vector.StrokeLine(screen,
		position[0], position[1]-halfSize,
		position[0], position[1]+halfSize,
		lineWidth, color, true)
}

// drawArrowHeadWithShader draws an arrowhead using triangles with the specified shader
func (r *Renderer) drawArrowHeadWithShader(screen *ebiten.Image, start, end f32.Vec2, color color.RGBA, shader *ebiten.Shader, uniforms map[string]any) {
	// Calculate arrow direction
	dx := end[0] - start[0]
	dy := end[1] - start[1]
	length := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	if length == 0 {
		return
	}

	// Normalize direction vector
	unitX := dx / length
	unitY := dy / length

	// Arrow head size
	arrowLength := float32(15.0)
	arrowWidth := float32(8.0)

	// Calculate perpendicular vector
	perpX := -unitY
	perpY := unitX

	// Calculate arrow head points
	backX := end[0] - unitX*arrowLength
	backY := end[1] - unitY*arrowLength

	leftX := backX + perpX*arrowWidth
	leftY := backY + perpY*arrowWidth

	rightX := backX - perpX*arrowWidth
	rightY := backY - perpY*arrowWidth

	// Draw arrow head lines with shader
	r.drawLineWithShader(screen, end, f32.Vec2{leftX, leftY}, 3, color, shader, uniforms)
	r.drawLineWithShader(screen, end, f32.Vec2{rightX, rightY}, 3, color, shader, uniforms)
}
