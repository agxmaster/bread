package colorize

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

const (
	ColorDraw  = "\x1b["
	ColorClean = "\x1b[0m"

	normalIntensityFg = 30
	highIntensityFg   = 90
	normalIntensityBg = 40
	highIntensityBg   = 100
)

// Where the magic happens
type Colorize struct {
	Values []interface{}
	Fg     Color
	Bg     Color
	Prop   Property
	plain  bool
}

func New(style string) Colorize {
	// Inpired from https://github.com/mgutz/ansi
	foreground2background := strings.Split(style, ":")
	foreground := strings.Split(foreground2background[0], "+")

	fg := colors[foreground[0]]
	fgStyle := ""
	if len(foreground) > 1 {
		fgStyle = foreground[1]
	}

	var bg Color
	bgStyle := ""
	if len(foreground2background) > 1 {
		background := strings.Split(foreground2background[1], "+")

		bg = colors[background[0]]
		if len(background) > 1 {
			bgStyle = background[1]
		}
	}

	c := Colorize{Fg: fg, Bg: bg, plain: false}
	if len(fgStyle) > 0 {
		if strings.Contains(fgStyle, "b") {
			c.ToggleBold()
		}
		if strings.Contains(fgStyle, "B") {
			c.ToggleBlink()
		}
		if strings.Contains(fgStyle, "u") {
			c.ToggleUnderline()
		}
		if strings.Contains(fgStyle, "i") {
			c.ToggleInverse()
		}
		if strings.Contains(fgStyle, "h") {
			c.ToggleFgIntensity()
		}
	}

	if len(bgStyle) > 0 {
		if strings.Contains(bgStyle, "h") {
			c.ToggleBgIntensity()
		}
	}

	return c
}

func (c *Colorize) SetPlain(plain bool) {
	c.plain = plain
}

func (c *Colorize) SetFgColor(fc Color) {
	c.Fg = fc
}

func (c *Colorize) SetBgColor(bc Color) {
	c.Bg = bc
}

func (c *Colorize) ToggleFgIntensity() {
	c.Prop.Fgi = !c.Prop.Fgi
}

func (c *Colorize) ToggleBgIntensity() {
	c.Prop.Bgi = !c.Prop.Bgi
}

func (c *Colorize) ToggleBold() {
	c.Prop.Bold = !c.Prop.Bold
}

func (c *Colorize) ToggleBlink() {
	c.Prop.Blink = !c.Prop.Blink
}

func (c *Colorize) ToggleUnderline() {
	c.Prop.Underline = !c.Prop.Underline
}

func (c *Colorize) ToggleInverse() {
	c.Prop.Inverse = !c.Prop.Inverse
}

func (c *Colorize) TogglePlain() {
	c.plain = !c.plain
}

// Paint returns colored string
func (c Colorize) Paint(args ...interface{}) string {
	c.Values = make([]interface{}, len(args))
	for i, arg := range args {
		switch arg.(type) {
		case string:
			c.Values[i] = strings.Replace(arg.(string), "%", "%%", -1)

		case []byte:
			c.Values[i] = strings.Replace(string(arg.([]byte)), "%", "%%", -1)

		default:
			c.Values[i] = arg

		}
	}

	return fmt.Sprint(c)
}

// Format allows Colorize to satisfy the fmt.Formatter interface.
func (c Colorize) Format(fs fmt.State, r rune) {
	begin, end := c.Colour()

	// start colour
	fmt.Fprint(fs, begin)

	max := len(c.Values) - 1
	for i, value := range c.Values {
		if i < max {
			fmt.Fprintf(fs, "%v ", value)
		} else {
			fmt.Fprintf(fs, fmt.Sprint(value))
		}

	}

	// clean state
	fmt.Fprint(fs, end)
}

func (c Colorize) Colour() (begin, end string) {
	if c.plain {
		return
	}

	var (
		base int
	)

	buf := bytes.NewBuffer(nil)

	// First Handle the Fg styles and options
	if c.Fg.HasColor() {
		if c.Prop.Fgi {
			base = highIntensityFg
		} else {
			base = normalIntensityFg
		}

		if c.Fg != ColorBlack {
			base += int(c.Fg)
		}

		buf.WriteString(ColorDraw)
		buf.WriteString("0;")
		buf.WriteString(c.Prop.String())
		buf.WriteString(strconv.Itoa(base))
		buf.WriteString("m")
	}

	// Next Handle the Bg styles and options
	if c.Bg.HasColor() {
		if c.Prop.Bgi {
			base = highIntensityBg
		} else {
			base = normalIntensityBg
		}

		if c.Bg != ColorBlack {
			base += int(c.Bg)
		}

		// We still want to honor props if only the background is set
		buf.WriteString(ColorDraw)

		if !c.Fg.HasColor() {
			buf.WriteString(c.Prop.String())
		}

		buf.WriteString(strconv.Itoa(base))
		buf.WriteString("m")
	}

	return buf.String(), ColorClean
}
