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

// drawCelestialBodies renders all celestial bodies with animated faces that track the player (EXCLUDING black holes)
func (r *Renderer) drawCelestialBodies(screen *ebiten.Image, model *Models.SpaceGame) {
	camera := model.Camera

	// First pass: draw all orbits (EXCLUDING black holes)
	for _, body := range model.CelestialBodies {
		// Skip black holes
		if body.GetType() == Models.CelestialBodyTypeBlackHole {
			continue
		}

		bodyPos := body.GetPosition()
		screenPos := camera.WorldToScreen(bodyPos, constants.ScreenWidth, constants.ScreenHeight)

		var orbitColor color.RGBA

		// Assign orbit color based on celestial body type
		switch body.GetType() {
		case Models.CelestialBodyTypePlanet:
			orbitColor = colors.PlanetOrbit
		case Models.CelestialBodyTypeWhiteHole:
			orbitColor = colors.WhiteHoleOrbit
		case Models.CelestialBodyTypeAsteroid:
			orbitColor = colors.AsteroidOrbit
		}

		orbitRadius := camera.RadiusToScreen(body.GetOrbitRadius(), constants.ScreenWidth, constants.ScreenHeight)
		r.drawOrbitCircleWithLight(screen, model, screenPos, orbitRadius, orbitColor)
	}

	// Second pass: draw all celestial bodies on top of orbits (EXCLUDING black holes)
	for _, body := range model.CelestialBodies {
		// Skip black holes
		if body.GetType() == Models.CelestialBodyTypeBlackHole {
			continue
		}

		bodyPos := body.GetPosition()
		screenPos := camera.WorldToScreen(bodyPos, constants.ScreenWidth, constants.ScreenHeight)
		radius := camera.RadiusToScreen(body.GetRadius(), constants.ScreenWidth, constants.ScreenHeight)

		var bodyColor color.RGBA

		// Assign body color based on celestial body type
		switch body.GetType() {
		case Models.CelestialBodyTypePlanet:
			bodyColor = color.RGBA{R: 255, G: 255, B: 255, A: 255} // White planet
		case Models.CelestialBodyTypeWhiteHole:
			bodyColor = color.RGBA{R: 255, G: 255, B: 255, A: 255} // White hole
		case Models.CelestialBodyTypeAsteroid:
			bodyColor = colors.AsteroidBodyAlt
		}

		// Use appropriate face shader for each celestial body type
		switch body.GetType() {
		case Models.CelestialBodyTypePlanet:
			if planet, ok := body.(*Models.Planet); ok {
				r.DrawPlanetWithFace(screen, model, planet)
			}
		case Models.CelestialBodyTypeWhiteHole:
			r.DrawWhiteHoleWithFace(screen, model, bodyPos, body.GetRadius(), body.GetOrbitRadius(), bodyColor)
		default:
			// Fallback to simple circle for other body types (like asteroids)
			vector.DrawFilledCircle(screen, screenPos[0], screenPos[1], radius, bodyColor, true)
		}
	}
}

// DrawBlackHoles renders ONLY black holes with their orbits and faces
func (r *Renderer) DrawBlackHoles(screen *ebiten.Image, model *Models.SpaceGame) {
	camera := model.Camera

	// First pass: draw black hole orbits
	for _, body := range model.CelestialBodies {
		// Only render black holes
		if body.GetType() != Models.CelestialBodyTypeBlackHole {
			continue
		}

		bodyPos := body.GetPosition()
		screenPos := camera.WorldToScreen(bodyPos, constants.ScreenWidth, constants.ScreenHeight)
		orbitColor := colors.BlackHoleOrbit
		orbitRadius := camera.RadiusToScreen(body.GetOrbitRadius(), constants.ScreenWidth, constants.ScreenHeight)

		r.drawOrbitCircleWithReveal(screen, model, screenPos, orbitRadius, orbitColor)
	}

	// Second pass: draw black hole bodies with faces
	for _, body := range model.CelestialBodies {
		// Only render black holes
		if body.GetType() != Models.CelestialBodyTypeBlackHole {
			continue
		}

		bodyPos := body.GetPosition()
		bodyColor := colors.BlackHoleBody

		r.DrawBlackHoleWithFace(screen, model, bodyPos, body.GetRadius(), body.GetOrbitRadius(), bodyColor)
	}
}

func (r *Renderer) DrawPlanetWithFace(screen *ebiten.Image, model *Models.SpaceGame, planet *Models.Planet) {
	screenPos := model.Camera.WorldToScreen(planet.Position, constants.ScreenWidth, constants.ScreenHeight)
	screenRadius := planet.GetRadius() * model.Camera.GetTotalZoom() * float32(constants.ScreenHeight)
	vector.DrawFilledCircle(screen, screenPos[0], screenPos[1], screenRadius, planet.BaseColor, true)

	if r.planetShader == nil {
		return
	}

	sw := float32(constants.ScreenWidth)
	sh := float32(constants.ScreenHeight)
	zoom := model.Camera.GetTotalZoom()
	t := float32(time.Since(r.startTime).Seconds())

	uniforms := map[string]any{
		"PlayerPos":         []float32{model.Player.Position[0], model.Player.Position[1]},
		"PlanetPos":         []float32{planet.Position[0], planet.Position[1]},
		"PlanetOrbitRadius": planet.OrbitRadius,
		"PlanetColor":       []float32{float32(planet.BaseColor.R), float32(planet.BaseColor.G), float32(planet.BaseColor.B), float32(planet.BaseColor.A)},
		"CameraPos":         []float32{model.Camera.Position[0], model.Camera.Position[1]},
		"Zoom":              zoom,
		"Radius":            planet.GetRadius(),
		"Time":              t,
		"ScreenSize":        []float32{sw, sh},
		"BaseColor":         []float32{float32(planet.BaseColor.R) / 255.0, float32(planet.BaseColor.G) / 255.0, float32(planet.BaseColor.B) / 255.0},
		"Seed":              planet.Seed,
	}

	// Draw directly to screen using the shader - no intermediate texture needed
	opts := &ebiten.DrawRectShaderOptions{Uniforms: uniforms}
	opts.Images[0] = r.whiteTexture // Use pre-allocated white texture as source
	screen.DrawRectShader(constants.ScreenWidth, constants.ScreenHeight, r.planetShader, opts)
}

// DrawBlackHoleWithFace renders a black hole with an evil facial expression that tracks the player
// The mouth grows larger and more menacing as the player gets closer to the orbit
func (r *Renderer) DrawBlackHoleWithFace(screen *ebiten.Image, model *Models.SpaceGame, blackHolePos f32.Vec2, worldRadius float32, worldOrbitRadius float32, blackHoleColor color.RGBA) {
	// First draw anti-aliased circle for base shape
	screenPos := model.Camera.WorldToScreen(blackHolePos, constants.ScreenWidth, constants.ScreenHeight)
	screenRadius := worldRadius * model.Camera.GetTotalZoom() * float32(constants.ScreenHeight)
	vector.DrawFilledCircle(screen, screenPos[0], screenPos[1], screenRadius, blackHoleColor, true)

	if r.blackHoleShader == nil {
		return
	}

	sw := float32(constants.ScreenWidth)
	sh := float32(constants.ScreenHeight)
	zoom := model.Camera.GetTotalZoom()
	t := float32(time.Since(r.startTime).Seconds())

	uniforms := map[string]any{
		"PlayerPos":            []float32{model.Player.Position[0], model.Player.Position[1]},
		"BlackHolePos":         []float32{blackHolePos[0], blackHolePos[1]},
		"BlackHoleOrbitRadius": worldOrbitRadius,
		"CameraPos":            []float32{model.Camera.Position[0], model.Camera.Position[1]},
		"Zoom":                 zoom,
		"Radius":               worldRadius,
		"Time":                 t,
		"ScreenSize":           []float32{sw, sh},
		//// Orbit overlap detection uniforms
		//"NumOtherOrbits":    numOtherOrbits,
		//"OtherOrbitCenters": otherOrbitCenters,
		//"OtherOrbitRadii":   otherOrbitRadii,
	}

	// Draw directly to screen using the shader - no intermediate texture needed
	opts := &ebiten.DrawRectShaderOptions{Uniforms: uniforms}
	opts.Images[0] = r.whiteTexture // Use pre-allocated white texture as source
	screen.DrawRectShader(constants.ScreenWidth, constants.ScreenHeight, r.blackHoleShader, opts)
}

// DrawWhiteHoleWithFace renders a white hole with a neutral facial expression that tracks the player
// The mouth subtly grows wider as the player approaches the orbit, maintaining a calm demeanor
func (r *Renderer) DrawWhiteHoleWithFace(screen *ebiten.Image, model *Models.SpaceGame, whiteHolePos f32.Vec2, worldRadius float32, worldOrbitRadius float32, whiteHoleColor color.RGBA) {
	// First draw anti-aliased circle for base shape
	screenPos := model.Camera.WorldToScreen(whiteHolePos, constants.ScreenWidth, constants.ScreenHeight)
	screenRadius := worldRadius * model.Camera.GetTotalZoom() * float32(constants.ScreenHeight)
	vector.DrawFilledCircle(screen, screenPos[0], screenPos[1], screenRadius, whiteHoleColor, true)

	if r.whiteHoleShader == nil {
		return
	}

	sw := float32(constants.ScreenWidth)
	sh := float32(constants.ScreenHeight)
	zoom := model.Camera.GetTotalZoom()
	t := float32(time.Since(r.startTime).Seconds())

	uniforms := map[string]any{
		"PlayerPos":            []float32{model.Player.Position[0], model.Player.Position[1]},
		"WhiteHolePos":         []float32{whiteHolePos[0], whiteHolePos[1]},
		"WhiteHoleOrbitRadius": worldOrbitRadius,
		"CameraPos":            []float32{model.Camera.Position[0], model.Camera.Position[1]},
		"Zoom":                 zoom,
		"Radius":               worldRadius,
		"Time":                 t,
		"ScreenSize":           []float32{sw, sh},
	}

	// Draw directly to screen using the shader - no intermediate texture needed
	opts := &ebiten.DrawRectShaderOptions{Uniforms: uniforms}
	opts.Images[0] = r.whiteTexture // Use pre-allocated white texture as source
	screen.DrawRectShader(constants.ScreenWidth, constants.ScreenHeight, r.whiteHoleShader, opts)
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

		// Collect all other celestial body orbit information for overlap detection
		otherOrbitCenters := make([]float32, 20) // up to 10 orbits * 2 coords
		otherOrbitRadii := make([]float32, 10)   // up to 10 orbits
		numOtherOrbits := 0

		// Gather all celestial body orbits except the current one
		for _, body := range model.CelestialBodies {
			if numOtherOrbits >= 10 {
				break // Limit to 10 orbits for performance
			}

			if body.GetType() == Models.CelestialBodyTypeBlackHole {
				continue // Skip black holes for orbit circles
			}

			bodyPos := body.GetPosition()
			bodyScreenPos := camera.WorldToScreen(bodyPos, constants.ScreenWidth, constants.ScreenHeight)
			bodyOrbitRadius := camera.RadiusToScreen(body.GetOrbitRadius(), constants.ScreenWidth, constants.ScreenHeight)

			// Debug: Print data for troubleshooting
			// fmt.Printf("DEBUG: Body at (%f,%f) with radius %f vs current (%f,%f) with radius %f\n",
			//     bodyScreenPos[0], bodyScreenPos[1], bodyOrbitRadius, center[0], center[1], radius)

			// Skip the current orbit circle we're rendering (avoid self-comparison)
			// Use more generous tolerance to avoid excluding valid orbits
			if math.Abs(float64(bodyScreenPos[0]-center[0])) < 5.0 &&
				math.Abs(float64(bodyScreenPos[1]-center[1])) < 5.0 &&
				math.Abs(float64(bodyOrbitRadius-radius)) < 5.0 {
				// fmt.Printf("DEBUG: Skipping self-orbit\n")
				continue
			}

			// Add this orbit to the list
			otherOrbitCenters[numOtherOrbits*2] = bodyScreenPos[0]   // x
			otherOrbitCenters[numOtherOrbits*2+1] = bodyScreenPos[1] // y
			otherOrbitRadii[numOtherOrbits] = bodyOrbitRadius
			numOtherOrbits++
			// fmt.Printf("DEBUG: Added other orbit %d at (%f,%f) with radius %f\n",
			//     numOtherOrbits-1, bodyScreenPos[0], bodyScreenPos[1], bodyOrbitRadius)
		}

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
			// Overlap detection uniforms
			"NumOtherOrbits":    numOtherOrbits,
			"OtherOrbitCenters": otherOrbitCenters,
			"OtherOrbitRadii":   otherOrbitRadii,
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
