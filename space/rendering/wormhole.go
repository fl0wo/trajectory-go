//go:build ignore
// +build ignore

//kage:unit pixels

package main

// Wormhole portal parameters
const PortalRimWidth = 0.02        // Width of the portal rim
const PortalGlowIntensity = 1.2    // Intensity of the glow effect
const PortalShimmerSpeed = 2.0     // Speed of the shimmer animation

// Uniforms from Go
var PortalA_Pos vec2       // [0..1] world-space portal A center
var PortalA_Radius float   // portal A radius in world units
var PortalA_Rotation float // portal A rotation in radians
var PortalA_Color vec3     // portal A color

var PortalB_Pos vec2       // [0..1] world-space portal B center  
var PortalB_Radius float   // portal B radius in world units
var PortalB_Rotation float // portal B rotation in radians
var PortalB_Color vec3     // portal B color

var CameraPos vec2         // [0..1] camera position
var Zoom float             // camera zoom level
var Time float             // current time for animations
var ScreenSize vec2        // [width, height] in pixels
var IsActive float         // 1.0 if portals are active, 0.0 if on cooldown

// Convert world position to screen coordinates
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

// Convert screen position to world coordinates
func screenToWorld(screenPos vec2) vec2 {
	uv := screenPos / ScreenSize
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

// Rotate a 2D vector by angle (in radians)
func rotate2D(v vec2, angle float) vec2 {
	c := cos(angle)
	s := sin(angle)
	return vec2(c*v.x - s*v.y, s*v.x + c*v.y)
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

// Sample the scene texture through a portal effect
func sampleThroughPortal(srcPos vec2, sourcePortalPos vec2, sourceRadius float, sourceRotation float, 
                        targetPortalPos vec2, targetRadius float, targetRotation float) vec4 {
	// Convert current screen position to world coordinates
	currentWorldPos := screenToWorld(srcPos)
	
	// Calculate distance from source portal center
	offsetFromSource := currentWorldPos - sourcePortalPos
	distFromSource := length(offsetFromSource)
	
	// If we're outside the source portal, return transparent
	if distFromSource > sourceRadius {
		return vec4(0.0, 0.0, 0.0, 0.0)
	}
	
	// Apply rotation difference between portals
	rotationDiff := targetRotation - sourceRotation
	rotatedOffset := rotate2D(offsetFromSource, rotationDiff)
	
	// Scale the offset to account for radius differences
	scale := targetRadius / sourceRadius
	scaledOffset := rotatedOffset * scale
	
	// Calculate target world position
	targetWorldPos := targetPortalPos + scaledOffset
	
	// Convert target world position back to screen coordinates
	targetScreenPos := worldToScreen(targetWorldPos)
	
	// Sample from the target location, but clamp to screen bounds
	targetUV := targetScreenPos / ScreenSize
	targetUV = clamp(targetUV, vec2(0.0), vec2(1.0))
	
	// Sample the scene texture
	sceneColor := imageSrc0At(targetUV * ScreenSize)
	
	// Apply portal mask (fade at edges)
	maskFactor := 1.0 - smoothstep(sourceRadius * 0.7, sourceRadius, distFromSource)
	
	return vec4(sceneColor.rgb, sceneColor.a * maskFactor)
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// Get current pixel's world position
	currentWorldPos := screenToWorld(dstPos.xy)
	
	// Check if pixel is inside Portal A
	distToA := length(currentWorldPos - PortalA_Pos)
	distToB := length(currentWorldPos - PortalB_Pos)
	
	// Use portal radii directly in world space
	worldRadiusA := PortalA_Radius
	worldRadiusB := PortalB_Radius
	
	// Sample base scene color
	baseColor := imageSrc0At(srcPos)
	finalColor := baseColor
	
	// Portal A effect - show Portal B's view
	if distToA <= worldRadiusA {
		portalColor := sampleThroughPortal(dstPos.xy, PortalA_Pos, worldRadiusA, PortalA_Rotation,
		                                  PortalB_Pos, worldRadiusB, PortalB_Rotation)
		
		// Add portal rim glow
		rimFactor := smoothstep(worldRadiusA - PortalRimWidth, worldRadiusA, distToA)
		rimColor := PortalA_Color * PortalGlowIntensity
		
		// Blend portal effect with rim
		if portalColor.a > 0.0 {
			finalColor = mix(baseColor, portalColor, portalColor.a * IsActive)
		}
		
		// Add rim glow
		finalColor.rgb = mix(finalColor.rgb, rimColor, rimFactor * 0.3 * IsActive)
	}
	
	// Portal B effect - show Portal A's view  
	if distToB <= worldRadiusB && distToA > worldRadiusA {
		portalColor := sampleThroughPortal(dstPos.xy, PortalB_Pos, worldRadiusB, PortalB_Rotation,
		                                  PortalA_Pos, worldRadiusA, PortalA_Rotation)
		
		// Add portal rim glow
		rimFactor := smoothstep(worldRadiusB - PortalRimWidth, worldRadiusB, distToB)
		rimColor := PortalB_Color * PortalGlowIntensity
		
		// Blend portal effect with rim
		if portalColor.a > 0.0 {
			finalColor = mix(baseColor, portalColor, portalColor.a * IsActive)
		}
		
		// Add rim glow
		finalColor.rgb = mix(finalColor.rgb, rimColor, rimFactor * 0.3 * IsActive)
	}
	
	// Add outer glow for both portals
	if distToA <= worldRadiusA * 1.5 {
		glowFactor := 1.0 - smoothstep(worldRadiusA, worldRadiusA * 1.5, distToA)
		finalColor.rgb = mix(finalColor.rgb, PortalA_Color, glowFactor * 0.1 * IsActive)
	}
	
	if distToB <= worldRadiusB * 1.5 {
		glowFactor := 1.0 - smoothstep(worldRadiusB, worldRadiusB * 1.5, distToB)
		finalColor.rgb = mix(finalColor.rgb, PortalB_Color, glowFactor * 0.1 * IsActive)
	}
	
	return finalColor
}