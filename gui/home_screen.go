package gui

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/you/trajectory/constants"
	"github.com/you/trajectory/gui/resources"
	"github.com/you/trajectory/space/colors"
	Models "github.com/you/trajectory/space/model"
	"log"
	"math"
	"time"
)

const nebulaShaderSource = `//kage:unit pixels

package main

// Uniforms
var Time float
var CameraPos vec2  // in normalized UV [0..1]
var ScreenSize vec2 // in pixels (only for initial UV calc)
var Zoom float      // <1 = zoomed out, >1 = zoomed in

// Pseudo-random [0..1)
func hash(p vec2) float {
	return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453)
}

// Zoom + camera parallax helper
func worldUV(uv vec2, parallax float) vec2 {
	uv = (uv-vec2(0.5, 0.5))/Zoom + vec2(0.5, 0.5)
	off := vec2(0.5, 0.5) - CameraPos
	return uv - off*parallax
}

func particles(uv vec2, layer float) vec4 {
	// per-layer tuning (unchanged)
	var scale, speed, bright, size, parallax, zc float
	if layer < 1.0 {
		scale, speed, bright, size, parallax, zc = 1.0, 0.3, 0.3, 0.6, 0.2, 0.1
	} else if layer < 2.0 {
		scale, speed, bright, size, parallax, zc = 2.0, 0.5, 0.6, 1.2, 0.3, 0.3
	} else {
		scale, speed, bright, size, parallax, zc = 3.0, 0.7, 1.0, 1.6, 0.5, 0.5
	}

	// common transforms (zoom + parallax)
	wuv := worldUV(uv, parallax)
	scaled := wuv * scale
	cell := floor(scaled)
	frac := fract(scaled)

	// per-cell randomness
	rX := hash(cell)
	rY := hash(cell + vec2(5.2, 1.3))
	t := hash(cell + vec2(8.7, 3.4))

	// base orbiting "wingle"
	ang := Time*speed + (rX+rY)*6.2831

	// compute a safe, inset center so wiggle never overflows cell
	radius := 0.008 * size / ((1 + Zoom) * zc)
	wiggleAmp := 0.0001 // ±0.10 UV units
	pad := radius + 0.2 + wiggleAmp
	jX := rX*(1.0-2.0*pad) + pad
	jY := rY*(1.0-2.0*pad) + pad
	basePos := vec2(jX, jY)

	// combine orbit + wiggle
	wiggleFreq := 2.0 // ≈3s per full cycle
	wiggle := vec2(
		sin(Time*wiggleFreq+rX*6.2831),
		cos(Time*wiggleFreq+rY*6.2831),
	) * wiggleAmp
	pos := fract(basePos + vec2(cos(ang), sin(ang))*0.2 + wiggle)

	// mask that circle
	local := frac - pos
	aspectInv := ScreenSize.y / ScreenSize.x
	localCorr := vec2(local.x, local.y*aspectInv)
	mask := step(length(localCorr), radius)

	// tiny opacity pulse
	pulse := 1.0 + 0.1*sin(Time*1.0+(rX+rY)*6.2831)

	// flat palette
	var col vec3
	if t < 0.02 {
		col = vec3(1.0, 1.0, 0.992)
	} else if t < 0.40 {
		col = vec3(0.129, 0.059, 0.314)
	} else if t < 0.75 {
		col = vec3(0.278, 0.102, 0.459)
	} else if t < 0.90 {
		col = vec3(0.282, 0.110, 0.459)
	} else if t < 0.96 {
		col = vec3(0.220, 0.008, 0.635)
	} else {
		col = vec3(0.031, 0.008, 0.106)
	}

	alpha := bright * mask * pulse
	return vec4(col, alpha)
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// normalized UV
	uv := (dstPos.xy - imageDstOrigin()) / ScreenSize

	// base rich-black background
	result := vec4(11/255.0, 2/255.0, 43/255.0, 1.0)

	// three animated circle layers
	for layer := 0.0; layer < 3.0; layer++ {
		p := particles(uv, layer)
		result.rgb += p.rgb * p.a
	}

	// clamp and return
	result.rgb = min(result.rgb, vec3(1.0))
	return result
}
`

// Scrolling behavior constants
const (
	DragHistoryCapacity   = 10
	DragHistoryTimeWindow = 0.1 // seconds

	ExponentialThreshold = 50.0  // pixels before exponential scaling kicks in
	ExponentialFactor    = 200.0 // pixels needed to double the effect

	ClickThreshold = 10.0 // pixels - below this is considered a click

	MomentumMultiplier = 6.0 // velocity to momentum conversion (reduced from 30.0)
	MomentumFriction   = 0.9 // friction applied each frame (reduced from 0.95)
	MinimumVelocity    = 1.0 // minimum velocity to keep momentum going (increased from 0.1)

	ScrollAnimationSpeed = 6.0        // scroll interpolation speed (reduced from 12.0)
	FallbackDeltaTime    = 1.0 / 60.0 // 60fps fallback

	BaseWheelScrollAmount = 40.0 // base wheel scroll amount (reduced from 80.0)
	WheelMultiplierFactor = 0.2  // wheel acceleration factor (reduced from 0.5)

	NodeSpacingMultiplier = 12.0 // spacing between level nodes
	NodeSizeMultiplier    = 0.8  // size of level nodes relative to button size

	// Level preview rectangles
	PreviewRectWidth  = 180.0 // width of level preview rectangles
	PreviewRectHeight = 120.0 // height of level preview rectangles

	// Preview image resolution (higher than display size for crisp quality)
	PreviewResolutionScale = 0.1                                  // percentage of screen width for preview resolution
	PreviewAspectRatio     = PreviewRectWidth / PreviewRectHeight // maintain aspect ratio

	LeftMarginDivisor     = 2.0 // divisor for left margin calculation
	TitleMarginMultiplier = 2.0 // title margin from top

	MagneticSnapDelay     = 0.1 // seconds to wait after user stops scrolling before snapping
	MagneticSnapSpeed     = 4.0 // speed of magnetic snapping animation
	MagneticSnapThreshold = 5.0 // pixels - how close to target before considering "snapped"

	CenterScalingMaxFactor = 2.0 // maximum scaling factor at center (3x bigger)
	CenterScalingRadius    = 3.0 // divisor for screen width to determine scaling zone (screen/4)

	CameraXOffsetScale = 2.0 // scale factor for camera X offset (reduced from 0.75)
)

type LevelNode struct {
	X, Y, Width, Height float32
	LevelNum            int
	IsLocked            bool
}

type HomeScreenImpl struct {
	starCount                                            int
	screenManager                                        *ScreenManager
	textFaceSource                                       *text.GoTextFaceSource
	layout                                               *ResponsiveLayout
	settingsButtonX, settingsButtonY, settingsButtonSize float32
	levelNodes                                           []LevelNode
	scrollOffsetX                                        float32 // Current horizontal scroll offset
	targetScrollX                                        float32 // Target scroll position for smooth scrolling
	isDragging                                           bool    // Whether user is currently dragging
	dragStartX                                           float32 // Mouse X position when drag started
	dragStartScrollX                                     float32 // Scroll offset when drag started
	totalPathWidth                                       float32 // Total width of the level path
	// Momentum scrolling variables
	scrollVelocity   float32   // Current scroll velocity
	lastDragX        float32   // Last drag position for velocity calculation
	dragHistory      []float32 // History of recent drag movements for momentum
	dragHistoryTimes []float32 // Timestamps for drag history
	lastFrameTime    float32   // Time of last frame for delta calculation

	// Magnetic snapping variables
	isSnapping           bool    // Whether currently performing magnetic snap
	snapTarget           float32 // Target position for magnetic snap
	snapTimer            float32 // Time since user stopped scrolling
	userStoppedScrolling bool    // Whether user has stopped all scrolling input

	// Nebula shader variables
	nebulaShader *ebiten.Shader
	startTime    time.Time

	// Level preview caching
	levelPreviews map[int]*ebiten.Image // Cache of rendered level previews
}

func (h *HomeScreenImpl) ResetScrollingState() {
	h.scrollVelocity = 0
	h.isDragging = false
	h.isSnapping = false
	h.dragHistory = h.dragHistory[:0]
	h.dragHistoryTimes = h.dragHistoryTimes[:0]
	h.userStoppedScrolling = false
	h.snapTimer = 0
	h.targetScrollX = h.scrollOffsetX
}

func NewHomeScreen(screenManager *ScreenManager) *HomeScreenImpl {
	// Create text face source
	textFaceSource, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal(err)
	}

	homeScreen := &HomeScreenImpl{
		starCount:            12, // Example star count
		screenManager:        screenManager,
		textFaceSource:       textFaceSource,
		layout:               NewResponsiveLayout(constants.ScreenWidth, constants.ScreenHeight), // Default, updated in Layout
		scrollOffsetX:        0,
		targetScrollX:        0,
		isDragging:           false,
		scrollVelocity:       0,
		dragHistory:          make([]float32, 0, DragHistoryCapacity), // Keep last N drag positions
		dragHistoryTimes:     make([]float32, 0, DragHistoryCapacity), // Keep last N timestamps
		lastFrameTime:        0,
		isSnapping:           false,
		snapTarget:           0,
		snapTimer:            0,
		userStoppedScrolling: false,
		startTime:            time.Now().Add(-10 * time.Second),
		levelPreviews:        make(map[int]*ebiten.Image),
	}

	// Initialize nebula shader
	homeScreen.initNebulaShader()

	homeScreen.generateLevelNodes()

	// Generate level previews
	homeScreen.generateLevelPreviews()

	return homeScreen
}

func (h *HomeScreenImpl) initNebulaShader() {
	shader, err := ebiten.NewShader([]byte(nebulaShaderSource))
	if err != nil {
		log.Printf("Failed to create nebula shader: %v", err)
		return
	}
	h.nebulaShader = shader
}

func (h *HomeScreenImpl) generateLevelPreviews() {
	// Get total number of levels
	totalLevels := len(Models.PredefinedLevels)

	// Load static preview images for each level
	for levelNum := 1; levelNum <= totalLevels; levelNum++ {
		previewImage := resources.LoadLevelPreview(levelNum)
		if previewImage != nil {
			h.levelPreviews[levelNum] = previewImage
		} else {
			log.Printf("Failed to load static preview for level %d", levelNum)
		}
	}
}

func (h *HomeScreenImpl) generateLevelNodes() {
	// Get total number of levels from PredefinedLevels
	totalLevels := len(Models.PredefinedLevels)
	h.levelNodes = make([]LevelNode, totalLevels)

	// For now, all levels are unlocked (you can add logic for locked levels later)
	for i := 1; i <= totalLevels; i++ {
		h.levelNodes[i-1] = LevelNode{
			LevelNum: i,
			IsLocked: false, // All unlocked for now
			Width:    0,     // Will be set in updateLevelNodePositions
			Height:   0,     // Will be set in updateLevelNodePositions
		}
	}
}

func (h *HomeScreenImpl) updateLevelNodePositions() {
	if len(h.levelNodes) == 0 {
		return
	}

	// Calculate responsive rectangle size and spacing
	width := float32(PreviewRectWidth)
	height := float32(PreviewRectHeight)
	spacing := width * 1.5 // Space between rectangle centers

	// Center vertically
	centerY := float32(h.layout.Height) / 2

	// Calculate total path width
	h.totalPathWidth = float32(len(h.levelNodes)-1) * spacing

	// Position each node horizontally with proper spacing
	for i := range h.levelNodes {
		h.levelNodes[i].X = float32(i) * spacing
		h.levelNodes[i].Y = centerY
		h.levelNodes[i].Width = width
		h.levelNodes[i].Height = height
	}
}

func (h *HomeScreenImpl) Update() error {
	// Update level node positions based on current layout
	h.updateLevelNodePositions()

	// Calculate delta time
	currentTime := float32(ebiten.TPS()) / 60.0 // Normalize to 60fps equivalent
	deltaTime := currentTime - h.lastFrameTime
	if deltaTime <= 0 {
		deltaTime = FallbackDeltaTime // Fallback to 60fps
	}
	h.lastFrameTime = currentTime

	// Handle mouse/touch input
	x, y := ebiten.CursorPosition()
	fx, fy := float32(x), float32(y)

	// Handle drag scrolling with momentum
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// Check if settings button was clicked first
		if IsPointInCircle(fx, fy, h.settingsButtonX+h.settingsButtonSize/2, h.settingsButtonY+h.settingsButtonSize/2, h.settingsButtonSize/2) {
			h.screenManager.SetScreen(SettingsScreen)
			return nil
		}

		// Start dragging
		h.isDragging = true
		h.dragStartX = fx
		h.lastDragX = fx
		h.dragStartScrollX = h.scrollOffsetX
		h.scrollVelocity = 0              // Stop any existing momentum
		h.dragHistory = h.dragHistory[:0] // Clear history
		h.dragHistoryTimes = h.dragHistoryTimes[:0]

		// Stop any magnetic snapping
		h.isSnapping = false
		h.snapTimer = 0
		h.userStoppedScrolling = false
	}

	if h.isDragging && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		// Calculate drag delta and velocity
		dragDelta := fx - h.lastDragX

		// Add to drag history for momentum calculation
		h.dragHistory = append(h.dragHistory, dragDelta)
		h.dragHistoryTimes = append(h.dragHistoryTimes, currentTime)

		// Keep only recent history (last N seconds)
		for len(h.dragHistory) > 0 && currentTime-h.dragHistoryTimes[0] > DragHistoryTimeWindow {
			h.dragHistory = h.dragHistory[1:]
			h.dragHistoryTimes = h.dragHistoryTimes[1:]
		}

		// Update scroll position with exponential scaling for fast drags
		totalDrag := fx - h.dragStartX

		// Exponential scaling: slow drags are 1:1, fast drags get amplified
		absTotal := float32(math.Abs(float64(totalDrag)))
		multiplier := float32(1.0)
		if absTotal > ExponentialThreshold {
			// Start exponential scaling after threshold pixels
			excess := absTotal - ExponentialThreshold
			multiplier = 1.0 + excess/ExponentialFactor // Each N px of excess doubles the effect
		}

		scaledDrag := totalDrag * multiplier
		h.targetScrollX = h.dragStartScrollX - scaledDrag // Reverse direction for natural scrolling

		h.lastDragX = fx

		// Clamp scroll position
		h.clampScrollPosition()
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		if h.isDragging {
			// Check if this was a click (small movement) rather than a drag
			dragDistance := float32(math.Abs(float64(fx - h.dragStartX)))
			if dragDistance < ClickThreshold { // Threshold for click vs drag
				// Check if any level node was clicked
				for _, node := range h.levelNodes {
					nodeScreenX := node.X - h.scrollOffsetX
					scaleFactor := h.calculateCenterScaling(nodeScreenX)
					scaledWidth := node.Width * scaleFactor
					scaledHeight := node.Height * scaleFactor
					// Rectangle coordinates (center-based to corner-based)
					rectX := nodeScreenX - scaledWidth/2
					rectY := node.Y - scaledHeight/2
					if !node.IsLocked && IsPointInRect(fx, fy, rectX, rectY, scaledWidth, scaledHeight) {
						// Load the selected level and go to game screen
						h.screenManager.SetScreenWithLevel(GameScreen, node.LevelNum)
						return nil
					}
				}
			} else {
				// Calculate momentum based on recent drag history
				if len(h.dragHistory) > 0 {
					var totalVelocity float32
					for _, delta := range h.dragHistory {
						totalVelocity += delta
					}
					// Average velocity over the history period
					avgVelocity := totalVelocity / float32(len(h.dragHistory))

					// Scale momentum based on drag intensity
					h.scrollVelocity = -avgVelocity * MomentumMultiplier // Reverse for natural scrolling
				}
			}
		}
		h.isDragging = false
	}

	// Apply momentum scrolling when not dragging
	if !h.isDragging && math.Abs(float64(h.scrollVelocity)) > MinimumVelocity {
		h.targetScrollX += h.scrollVelocity * deltaTime
		h.clampScrollPosition()

		// Apply friction to gradually slow down momentum
		h.scrollVelocity *= MomentumFriction
	}

	// Smooth scrolling animation towards target
	t := ScrollAnimationSpeed * deltaTime
	if t > 1.0 {
		t = 1.0
	}
	h.scrollOffsetX += (h.targetScrollX - h.scrollOffsetX) * t

	// Handle mouse wheel scrolling with exponential scaling
	_, wheelY := ebiten.Wheel()
	if wheelY != 0 {
		// Exponential wheel scrolling - multiple spins = faster scrolling
		baseAmount := float32(wheelY) * -BaseWheelScrollAmount
		wheelMultiplier := float32(math.Abs(float64(wheelY))) // More spins = higher multiplier
		scrollAmount := baseAmount * (1.0 + wheelMultiplier*WheelMultiplierFactor)

		h.targetScrollX += scrollAmount
		h.scrollVelocity = 0 // Stop momentum when using wheel
		h.clampScrollPosition()

		// Stop any magnetic snapping when user uses wheel
		h.isSnapping = false
		h.snapTimer = 0
		h.userStoppedScrolling = false
	}

	// Magnetic snapping logic
	h.updateMagneticSnapping(deltaTime)

	return nil
}

func (h *HomeScreenImpl) clampScrollPosition() {
	layoutWidth := float32(h.layout.Width)
	maxScroll := h.totalPathWidth - layoutWidth/LeftMarginDivisor
	if maxScroll < 0 {
		maxScroll = 0
	}

	if h.targetScrollX < -layoutWidth/LeftMarginDivisor {
		h.targetScrollX = -layoutWidth / LeftMarginDivisor // Allow some space on the left
		h.scrollVelocity = 0                               // Stop momentum at boundaries
	} else if h.targetScrollX > maxScroll {
		h.targetScrollX = maxScroll
		h.scrollVelocity = 0 // Stop momentum at boundaries
	}
}

// findClosestLevelToCenter finds the level node closest to the center of the screen
func (h *HomeScreenImpl) findClosestLevelToCenter() float32 {
	if len(h.levelNodes) == 0 {
		return 0
	}

	screenCenter := float32(h.layout.Width) / 2

	closestDistance := float32(math.MaxFloat32)
	closestNodeX := h.levelNodes[0].X

	for _, node := range h.levelNodes {
		// Calculate where this node appears on screen
		nodeScreenX := node.X - h.scrollOffsetX

		// Distance from screen center
		distance := float32(math.Abs(float64(nodeScreenX - screenCenter)))

		if distance < closestDistance {
			closestDistance = distance
			closestNodeX = node.X
		}
	}

	// Calculate the scroll position that would center this node
	return closestNodeX - screenCenter
}

// updateMagneticSnapping handles the magnetic snapping effect to center levels
func (h *HomeScreenImpl) updateMagneticSnapping(deltaTime float32) {
	// Check if user has stopped all scrolling input
	userActive := h.isDragging || math.Abs(float64(h.scrollVelocity)) > MinimumVelocity

	if userActive {
		// User is still actively scrolling
		h.userStoppedScrolling = false
		h.snapTimer = 0
		h.isSnapping = false
		return
	}

	// User has stopped scrolling
	if !h.userStoppedScrolling {
		h.userStoppedScrolling = true
		h.snapTimer = 0
		h.isSnapping = false
	}

	// Increment timer
	h.snapTimer += deltaTime

	// Start snapping after delay
	if !h.isSnapping && h.snapTimer >= MagneticSnapDelay {
		h.isSnapping = true
		h.snapTarget = h.findClosestLevelToCenter()
	}

	// Perform magnetic snapping animation
	if h.isSnapping {
		// Calculate distance to target
		distance := h.snapTarget - h.targetScrollX

		// Check if we're close enough to consider snapped
		if math.Abs(float64(distance)) < MagneticSnapThreshold {
			h.targetScrollX = h.snapTarget
			h.isSnapping = false
		} else {
			// Smooth interpolation towards target
			snapSpeed := MagneticSnapSpeed * deltaTime
			if snapSpeed > 1.0 {
				snapSpeed = 1.0
			}
			h.targetScrollX += distance * snapSpeed
		}
	}
}

// calculateCenterScaling calculates the scaling factor for a level node based on its distance from screen center
func (h *HomeScreenImpl) calculateCenterScaling(nodeScreenX float32) float32 {
	screenCenter := float32(h.layout.Width) / 2
	scalingZone := float32(h.layout.Width) / CenterScalingRadius // screen/4

	// Calculate distance from center
	distanceFromCenter := float32(math.Abs(float64(nodeScreenX - screenCenter)))

	// If outside scaling zone, no scaling
	if distanceFromCenter > scalingZone {
		return 1.0
	}

	// Linear interpolation: 0 distance = 3x scale, scalingZone distance = 1x scale
	normalizedDistance := distanceFromCenter / scalingZone
	scaleFactor := CenterScalingMaxFactor - (CenterScalingMaxFactor-1.0)*normalizedDistance

	return scaleFactor
}

func (h *HomeScreenImpl) Draw(screen *ebiten.Image) {
	// Draw nebula background shader if available
	if h.nebulaShader != nil {
		h.drawNebulaBackground(screen)
	} else {
		// Fallback to solid color background
		screen.Fill(colors.BorderPreviewLevel)
	}

	margin := h.layout.GetMargin()

	// Draw star count in top-left corner
	starText := fmt.Sprintf("S %d", h.starCount)
	starFace := &text.GoTextFace{
		Source: h.textFaceSource,
		Size:   h.layout.GetBodyFontSize(),
	}

	starOp := &text.DrawOptions{}
	starOp.GeoM.Translate(float64(margin), float64(margin))
	starOp.ColorScale.ScaleWithColor(colors.BorderPreviewLevel)
	text.Draw(screen, starText, starFace, starOp)

	// Draw settings icon in top-right corner
	h.settingsButtonSize = float32(h.layout.GetButtonSize())
	h.settingsButtonX = float32(h.layout.Width-margin) - h.settingsButtonSize
	h.settingsButtonY = float32(margin)

	DrawSettingsIcon(screen, h.settingsButtonX, h.settingsButtonY, h.settingsButtonSize, colors.BorderPreviewLevel)

	// Draw game title at top
	titleText := "Throw The Alien"
	titleFace := &text.GoTextFace{
		Source: h.textFaceSource,
		Size:   h.layout.GetTitleFontSize(),
	}

	titleWidthF, _ := text.Measure(titleText, titleFace, 0)
	titleWidth := int(titleWidthF)

	titleX := (h.layout.Width - titleWidth) / 2
	titleY := int(float32(margin) * TitleMarginMultiplier)

	titleOp := &text.DrawOptions{}
	titleOp.GeoM.Translate(float64(titleX), float64(titleY))
	titleOp.ColorScale.ScaleWithColor(colors.BorderPreviewLevel)
	text.Draw(screen, titleText, titleFace, titleOp)

	// Draw connecting lines between level nodes first (so they appear behind rectangles)
	lineColor := colors.BorderPreviewLevel // Light blue connecting lines
	for i := 0; i < len(h.levelNodes)-1; i++ {
		currentNode := h.levelNodes[i]
		nextNode := h.levelNodes[i+1]

		// Calculate screen positions (accounting for scroll)
		currentX := currentNode.X - h.scrollOffsetX
		nextX := nextNode.X - h.scrollOffsetX

		// Only draw lines that are visible on screen
		if (currentX > -currentNode.Width && currentX < float32(h.layout.Width)+currentNode.Width) ||
			(nextX > -nextNode.Width && nextX < float32(h.layout.Width)+nextNode.Width) {
			DrawConnectionLine(screen, currentX, currentNode.Y, nextX, nextNode.Y, lineColor)
		}
	}

	// Draw level nodes (rectangles with level previews)
	levelNumberFace := &text.GoTextFace{
		Source: h.textFaceSource,
		Size:   h.layout.GetBodyFontSize(),
	}

	for _, node := range h.levelNodes {
		// Calculate screen position (accounting for scroll)
		screenX := node.X - h.scrollOffsetX

		// Calculate scaling factor based on distance from center
		scaleFactor := h.calculateCenterScaling(screenX)
		scaledWidth := node.Width * scaleFactor
		scaledHeight := node.Height * scaleFactor

		// Only draw nodes that are visible on screen (with some margin)
		if screenX > -scaledWidth && screenX < float32(h.layout.Width)+scaledWidth {
			// Get the cached preview image for this level
			previewImage := h.levelPreviews[node.LevelNum]
			DrawLevelPreviewRect(screen, screenX, node.Y, scaledWidth, scaledHeight, node.LevelNum, levelNumberFace, node.IsLocked, previewImage)
		}
	}

	// Draw scroll instructions at bottom
	var instructionText string
	if h.layout.IsMobile() {
		instructionText = "Drag to scroll • Tap circle to play level"
	} else {
		instructionText = "Drag to scroll • Click circle to play level • Mouse wheel to scroll"
	}

	instructionFace := &text.GoTextFace{
		Source: h.textFaceSource,
		Size:   h.layout.GetSmallFontSize(),
	}

	instructionWidthF, _ := text.Measure(instructionText, instructionFace, 0)
	instructionWidth := int(instructionWidthF)
	instructionX := (h.layout.Width - instructionWidth) / 2
	instructionY := h.layout.Height - margin - int(h.layout.GetSmallFontSize())

	instructionOp := &text.DrawOptions{}
	instructionOp.GeoM.Translate(float64(instructionX), float64(instructionY))
	instructionOp.ColorScale.ScaleWithColor(colors.BorderPreviewLevel)
	text.Draw(screen, instructionText, instructionFace, instructionOp)
}

func (h *HomeScreenImpl) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, geoM ebiten.GeoM) {
	screen.DrawImage(offscreen, &ebiten.DrawImageOptions{
		GeoM: geoM,
	})
}

func (h *HomeScreenImpl) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	h.layout = NewResponsiveLayout(outsideWidth, outsideHeight)
	return outsideWidth, outsideHeight
}

func (h *HomeScreenImpl) SetStarCount(count int) {
	h.starCount = count
}

func (h *HomeScreenImpl) drawNebulaBackground(screen *ebiten.Image) {
	// Calculate time since start
	elapsed := time.Since(h.startTime).Seconds()

	// Get screen dimensions
	width := float32(h.layout.Width)
	height := float32(h.layout.Height)

	// Calculate camera position based on scroll offset
	// Normalize scroll position to [0,1] range for shader
	cameraPosX := h.scrollOffsetX / (h.totalPathWidth + width) * CameraXOffsetScale
	//if cameraPosX < 0 {
	//	cameraPosX = 0
	//} else if cameraPosX > 1 {
	//	cameraPosX = 1
	//}
	cameraPosY := float32(0.5) // Center vertically

	// Create shader options with uniforms
	opts := &ebiten.DrawTrianglesShaderOptions{
		Uniforms: map[string]interface{}{
			"Time":       float32(elapsed),
			"CameraPos":  []float32{cameraPosX, cameraPosY},
			"ScreenSize": []float32{float32(constants.ScreenWidth), float32(constants.ScreenHeight)},
			"Zoom":       float32(1.0), // Fixed zoom as requested
		},
	}

	// Create vertices for full screen quad
	vertices := []ebiten.Vertex{
		{DstX: 0, DstY: 0, SrcX: 0, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: width, DstY: 0, SrcX: 0, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: 0, DstY: height, SrcX: 0, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: width, DstY: height, SrcX: 0, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
	}

	// Create indices for two triangles forming a quad
	indices := []uint16{0, 1, 2, 1, 2, 3}

	// Draw the shader
	screen.DrawTrianglesShader(vertices, indices, h.nebulaShader, opts)
}
