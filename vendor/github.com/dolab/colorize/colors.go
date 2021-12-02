package colorize

const (
	ColorNone = iota
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
	ColorGray
	ColorBlack Color = -1
)

var (
	colors = map[string]Color{
		"red":     ColorRed,
		"green":   ColorGreen,
		"yellow":  ColorYellow,
		"blue":    ColorBlue,
		"magenta": ColorMagenta,
		"cyan":    ColorCyan,
		"white":   ColorWhite,
		"gray":    ColorGray,
		"black":   ColorBlack,
	}
)

// Color can be used for foreground and background
type Color int

func (c Color) HasColor() bool {
	return c != ColorNone
}
