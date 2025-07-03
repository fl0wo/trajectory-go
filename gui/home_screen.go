package gui

import (
	"bytes"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	Models "github.com/you/trajectory/space/model"
	"image/color"
	"log"
	"math"
)

// Scrolling behavior constants
const (
	// Drag history tracking
	DragHistoryCapacity   = 10
	DragHistoryTimeWindow = 0.1 // seconds

	// Exponential scaling
	ExponentialThreshold = 50.0  // pixels before exponential scaling kicks in
	ExponentialFactor    = 200.0 // pixels needed to double the effect

	// Click vs drag detection
	ClickThreshold = 10.0 // pixels - below this is considered a click

	// Momentum and friction
	MomentumMultiplier = 8.0  // velocity to momentum conversion (reduced from 30.0)
	MomentumFriction   = 0.92 // friction applied each frame (reduced from 0.95)
	MinimumVelocity    = 1.0  // minimum velocity to keep momentum going (increased from 0.1)

	// Animation speeds
	ScrollAnimationSpeed = 6.0        // scroll interpolation speed (reduced from 12.0)
	FallbackDeltaTime    = 1.0 / 60.0 // 60fps fallback

	// Mouse wheel scrolling
	BaseWheelScrollAmount = 40.0 // base wheel scroll amount (reduced from 80.0)
	WheelMultiplierFactor = 0.2  // wheel acceleration factor (reduced from 0.5)

	// Level node spacing
	NodeSpacingMultiplier = 8.0 // spacing between level nodes (reduced from 10.0)
	NodeSizeMultiplier    = 0.8 // size of level nodes relative to button size

	// Layout margins
	LeftMarginDivisor     = 2.0 // divisor for left margin calculation
	TitleMarginMultiplier = 2.0 // title margin from top
)

type LevelNode struct {
	X, Y, Radius float32
	LevelNum     int
	IsLocked     bool
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
}

func NewHomeScreen(screenManager *ScreenManager) *HomeScreenImpl {
	// Create text face source
	textFaceSource, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal(err)
	}

	homeScreen := &HomeScreenImpl{
		starCount:        12, // Example star count
		screenManager:    screenManager,
		textFaceSource:   textFaceSource,
		layout:           NewResponsiveLayout(1920, 1080), // Default, updated in Layout
		scrollOffsetX:    0,
		targetScrollX:    0,
		isDragging:       false,
		scrollVelocity:   0,
		dragHistory:      make([]float32, 0, DragHistoryCapacity), // Keep last N drag positions
		dragHistoryTimes: make([]float32, 0, DragHistoryCapacity), // Keep last N timestamps
		lastFrameTime:    0,
	}

	homeScreen.generateLevelNodes()
	return homeScreen
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
			Radius:   0,     // Will be set in updateLevelNodePositions
		}
	}
}

func (h *HomeScreenImpl) updateLevelNodePositions() {
	if len(h.levelNodes) == 0 {
		return
	}

	// Calculate responsive circle size and spacing
	radius := float32(h.layout.GetButtonSize()) * NodeSizeMultiplier
	spacing := radius * NodeSpacingMultiplier // Space between circle centers

	// Center vertically
	centerY := float32(h.layout.Height) / 2

	// Calculate total path width
	h.totalPathWidth = float32(len(h.levelNodes)-1) * spacing

	// Position each node horizontally with proper spacing
	for i := range h.levelNodes {
		h.levelNodes[i].X = float32(i) * spacing
		h.levelNodes[i].Y = centerY
		h.levelNodes[i].Radius = radius
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
					if !node.IsLocked && IsPointInCircle(fx, fy, nodeScreenX, node.Y, node.Radius) {
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
	}

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

func (h *HomeScreenImpl) Draw(screen *ebiten.Image) {
	// Fill screen with dark blue background (space-like)
	screen.Fill(color.RGBA{15, 25, 50, 255})

	margin := h.layout.GetMargin()

	// Draw star count in top-left corner
	starText := fmt.Sprintf("⭐ %d", h.starCount)
	starFace := &text.GoTextFace{
		Source: h.textFaceSource,
		Size:   h.layout.GetBodyFontSize(),
	}

	starOp := &text.DrawOptions{}
	starOp.GeoM.Translate(float64(margin), float64(margin))
	starOp.ColorScale.ScaleWithColor(color.RGBA{255, 215, 0, 255}) // Gold color
	text.Draw(screen, starText, starFace, starOp)

	// Draw settings icon in top-right corner
	h.settingsButtonSize = float32(h.layout.GetButtonSize())
	h.settingsButtonX = float32(h.layout.Width-margin) - h.settingsButtonSize
	h.settingsButtonY = float32(margin)

	DrawSettingsIcon(screen, h.settingsButtonX, h.settingsButtonY, h.settingsButtonSize, color.RGBA{255, 255, 255, 255})

	// Draw game title at top
	titleText := "TRAJECTORY"
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
	titleOp.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, 255})
	text.Draw(screen, titleText, titleFace, titleOp)

	// Draw connecting lines between level nodes first (so they appear behind circles)
	lineColor := color.RGBA{100, 150, 255, 180} // Light blue connecting lines
	for i := 0; i < len(h.levelNodes)-1; i++ {
		currentNode := h.levelNodes[i]
		nextNode := h.levelNodes[i+1]

		// Calculate screen positions (accounting for scroll)
		currentX := currentNode.X - h.scrollOffsetX
		nextX := nextNode.X - h.scrollOffsetX

		// Only draw lines that are visible on screen
		if (currentX > -currentNode.Radius && currentX < float32(h.layout.Width)+currentNode.Radius) ||
			(nextX > -nextNode.Radius && nextX < float32(h.layout.Width)+nextNode.Radius) {
			DrawConnectionLine(screen, currentX, currentNode.Y, nextX, nextNode.Y, lineColor)
		}
	}

	// Draw level nodes (circles with numbers)
	levelNumberFace := &text.GoTextFace{
		Source: h.textFaceSource,
		Size:   h.layout.GetBodyFontSize(),
	}

	for _, node := range h.levelNodes {
		// Calculate screen position (accounting for scroll)
		screenX := node.X - h.scrollOffsetX

		// Only draw nodes that are visible on screen (with some margin)
		if screenX > -node.Radius*2 && screenX < float32(h.layout.Width)+node.Radius*2 {
			DrawLevelCircle(screen, screenX, node.Y, node.Radius, node.LevelNum, levelNumberFace, node.IsLocked)
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
	instructionOp.ColorScale.ScaleWithColor(color.RGBA{128, 128, 128, 255})
	text.Draw(screen, instructionText, instructionFace, instructionOp)
}

func (h *HomeScreenImpl) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	h.layout = NewResponsiveLayout(outsideWidth, outsideHeight)
	return outsideWidth, outsideHeight
}

func (h *HomeScreenImpl) SetStarCount(count int) {
	h.starCount = count
}
