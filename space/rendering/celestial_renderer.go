package rendering

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/you/trajectory/constants"
	"github.com/you/trajectory/space/colors"
	Models "github.com/you/trajectory/space/model"
	"github.com/you/trajectory/space/resources"
	"github.com/you/trajectory/space/util"
	"golang.org/x/image/math/f32"
	"image/color"
	"math"
	"time"
)

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
		orbitRadius := camera.RadiusToScreen(body.GetOrbitRadius(), constants.ScreenWidth, constants.ScreenHeight)

		// Use reveal shader for black holes, regular light effect for others
		if body.GetType() == Models.CelestialBodyTypeBlackHole {
			r.drawOrbitCircleWithReveal(screen, model, screenPos, orbitRadius, orbitColor)
		} else {
			r.drawOrbitCircleWithLight(screen, model, screenPos, orbitRadius, orbitColor)
		}
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
			"Zoom":           camera.GetTotalZoom(),
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

// drawOrbitCircleWithReveal draws a dashed orbit circle that's only visible in the light area (for black holes)
func (r *Renderer) drawOrbitCircleWithReveal(screen *ebiten.Image, model *Models.SpaceGame, center f32.Vec2, radius float32, orbitColor color.RGBA) {
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

	// If shadows are enabled and we have the reveal shader, apply reveal effect
	if model.ShadowsEnabled && r.revealOnLightShader != nil {
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

		// Prepare shader uniforms using the reveal shader helper
		uniforms := r.prepareRevealOnLightUniforms(
			lightPos, lightDirVec,
			float32(fov*math.Pi/180.0), // Convert to radians
			float32(maxDistance),
			camera.GetTotalZoom(),
			orbitColor,
		)

		// Additional uniforms for the dashed circle animation
		uniforms["Time"] = currentTime
		uniforms["RotationDirection"] = float32(1.0) // Counterclockwise rotation
		uniforms["CircleCenter"] = []float32{center[0], center[1]}
		uniforms["CircleRadius"] = radius

		// Use shader-enabled dashed circle with reveal effect
		util.StrokeDashedCircleTrianglesWithShader(screen, center[0], center[1], radius, 4, orbitColor, dashLen, gapLen, true, r.revealOnLightShader, uniforms, currentTime/10.0)
	} else {
		// Fallback: don't draw anything (black holes should be invisible without light)
		// This creates the effect that black hole orbits are only visible when illuminated
	}
}
