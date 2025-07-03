package gui

import "math"

type ResponsiveLayout struct {
	Width  int
	Height int
	DPI    float64
}

func NewResponsiveLayout(width, height int) *ResponsiveLayout {
	return &ResponsiveLayout{
		Width:  width,
		Height: height,
		DPI:    1.0, // Default DPI
	}
}

func (r *ResponsiveLayout) IsPortrait() bool {
	return r.Height > r.Width
}

func (r *ResponsiveLayout) IsLandscape() bool {
	return r.Width > r.Height
}

func (r *ResponsiveLayout) IsMobile() bool {
	// Consider mobile if smaller dimension is less than 768px
	minDimension := int(math.Min(float64(r.Width), float64(r.Height)))
	return minDimension < 768
}

func (r *ResponsiveLayout) GetScaleFactor() float64 {
	// Base scale on smaller dimension for consistent sizing
	minDimension := math.Min(float64(r.Width), float64(r.Height))
	baseSize := 1080.0 // Reference size
	return minDimension / baseSize
}

func (r *ResponsiveLayout) ScaledSize(baseSize float64) float64 {
	return baseSize * r.GetScaleFactor()
}

func (r *ResponsiveLayout) GetMargin() int {
	if r.IsMobile() {
		return int(r.ScaledSize(20))
	}
	return int(r.ScaledSize(40))
}

func (r *ResponsiveLayout) GetButtonSize() int {
	if r.IsMobile() {
		return int(r.ScaledSize(80))
	}
	return int(r.ScaledSize(60))
}

func (r *ResponsiveLayout) GetTitleFontSize() float64 {
	if r.IsMobile() {
		return r.ScaledSize(48 * 2)
	}
	return r.ScaledSize(64 * 2)
}

func (r *ResponsiveLayout) GetBodyFontSize() float64 {
	if r.IsMobile() {
		return r.ScaledSize(48)
	}
	return r.ScaledSize(64)
}

func (r *ResponsiveLayout) GetSmallFontSize() float64 {
	if r.IsMobile() {
		return r.ScaledSize(18)
	}
	return r.ScaledSize(24)
}
