//go:build ignore
// +build ignore

//kage:unit pixels

package main

// Portal parameters
const PortalRimWidth = 0.08     // Width of the glowing rim
const PortalCornerRadius = 0.25 // Radius of the rounded corners
const PortalGlowIntensity = 1.5 // Intensity of the glow effect
const PortalShimmerSpeed = 3.0  // Speed of the shimmer animation

// Uniforms from Go
var PortalPos vec2       // [0..1] world-space portal center
var PortalWidth float    // portal width in world units
var PortalHeight float   // portal height in world units
var PortalRotation float // portal rotation in radians
var CameraPos vec2       // [0..1] camera position
var Zoom float           // camera zoom level
var Time float           // current time for animations
var ScreenSize vec2      // [width, height] in pixels
var IsActive float       // 1.0 if active, 0.0 if on cooldown
var CooldownTimer float  // cooldown timer (for dimming effect)
var PortalColor vec3     // base color of the portal

// SDF for a rounded rectangle (capsule shape)
func sdfRoundedRect(p vec2, size vec2, r float) float {
	d := abs(p) - size + r
	return min(max(d.x, d.y), 0.0) + length(max(d, 0.0)) - r
}

// Map screen position to world coordinates
func screenToWorld(dstPos vec4) vec2 {
	uv := dstPos.xy / ScreenSize
	c := uv - vec2(0.5)

	// Apply aspect ratio correction
	asp := ScreenSize.x / ScreenSize.y
	c.x *= asp

	// Apply zoom
	zoomSafe := max(0.01, min(Zoom, 100.0))
	c /= zoomSafe

	// Convert to world coordinates
	return CameraPos + c
}

// Convert world position to screen coordinates (inverse of screenToWorld)
func worldToScreen(worldPos vec2) vec2 {
	// Get relative position from camera
	relativePos := worldPos - CameraPos

	// Apply zoom
	zoomSafe := max(0.01, min(Zoom, 100.0))
	relativePos *= zoomSafe

	// Apply aspect ratio correction
	asp := ScreenSize.x / ScreenSize.y
	relativePos.x /= asp

	// Convert to screen coordinates
	uv := relativePos + vec2(0.5)
	return uv * ScreenSize
}

// Rotate a 2D vector by angle (in radians)
func rotate2D(v vec2, angle float) vec2 {
	c := cos(angle)
	s := sin(angle)
	return vec2(c*v.x-s*v.y, s*v.x+c*v.y)
}

// Generate noise for shimmer effect
func noise(p vec2) float {
	return fract(sin(dot(p, vec2(12.9898, 78.233))) * 43758.5453)
}

// Smoothstep function for smooth transitions
func smoothstep(edge0, edge1, x float) float {
	t := clamp((x-edge0)/(edge1-edge0), 0.0, 1.0)
	return t * t * (3.0 - 2.0*t)
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// Convert portal world position to screen coordinates
	portalScreenPos := worldToScreen(PortalPos)

	// Get current pixel position
	currentScreenPos := dstPos.xy

	// Calculate offset from portal center in screen space
	offsetPixels := currentScreenPos - portalScreenPos

	// Convert portal dimensions to screen pixels
	zoomSafe := max(0.01, min(Zoom, 100.0))
	portalScreenWidth := PortalWidth * zoomSafe * ScreenSize.y
	portalScreenHeight := PortalHeight * zoomSafe * ScreenSize.y

	// Convert to normalized portal-local coordinates
	localPos := vec2(offsetPixels.x/(portalScreenWidth*0.5), offsetPixels.y/(portalScreenHeight*0.5))

	// Apply rotation
	localPos = rotate2D(localPos, -PortalRotation)

	// Portal dimensions in local space (normalized)
	halfWidth := 1.0
	halfHeight := portalScreenHeight / portalScreenWidth

	// Create capsule shape (rounded rectangle)
	cornerRadius := min(halfWidth, halfHeight) * PortalCornerRadius
	portalSize := vec2(halfWidth-cornerRadius, halfHeight-cornerRadius)

	// Distance to portal edge
	distToPortal := sdfRoundedRect(localPos, portalSize, cornerRadius)

	// Portal rim (outer glow)
	rimDistance := abs(distToPortal) - PortalRimWidth
	rimAlpha := 1.0 - smoothstep(0.0, PortalRimWidth, abs(rimDistance))

	// Inner portal area
	innerAlpha := 1.0 - smoothstep(-0.02, 0.0, distToPortal)

	// Shimmer effect
	shimmerTime := Time * PortalShimmerSpeed
	shimmerNoise := noise(localPos*10.0 + shimmerTime)
	shimmerIntensity := 0.3 + 0.2*sin(shimmerTime+shimmerNoise*6.28318)

	// Activity state (dimming when on cooldown)
	activityLevel := IsActive
	if CooldownTimer > 0.0 {
		activityLevel *= 0.3 + 0.7*(1.0-CooldownTimer)
	}

	// Combine rim glow and inner portal
	finalAlpha := max(rimAlpha, innerAlpha) * activityLevel

	// Color calculation
	baseColor := PortalColor
	glowColor := baseColor * PortalGlowIntensity

	// Add shimmer to the color
	finalColor := mix(baseColor, glowColor, shimmerIntensity)

	// Fade out at edges for smooth blending
	edgeFade := 1.0 - smoothstep(max(halfWidth, halfHeight)*0.8, max(halfWidth, halfHeight)*1.2, length(localPos))
	finalAlpha *= edgeFade

	return vec4(finalColor, finalAlpha)
}
