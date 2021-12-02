package colorize

const (
	bold      = "1;"
	blink     = "5;"
	underline = "4;"
	inverse   = "7;"
)

// Can set 1 or more of these properties
// This struct holds the state
type Property struct {
	Bold      bool
	Blink     bool
	Underline bool
	Inverse   bool
	Fgi       bool
	Bgi       bool
}

func (p Property) String() string {
	var ppt string
	if p.Bold {
		ppt += bold
	}
	if p.Blink {
		ppt += blink
	}
	if p.Underline {
		ppt += underline
	}
	if p.Inverse {
		ppt += inverse
	}

	return ppt
}
