package common

type Header map[string][]string

// Len returns the number of items in h.
func (h Header) Len() int {
	return len(h)
}

// Copy returns a copy of h.
func (h Header) Copy() Header {
	return Join(h)
}

// Get obtains the values for a given key.
func (h Header) Get(k string) []string {
	return h[k]
}

// Set sets the value of a given key with a slice of values.
func (h Header) Set(k string, vals ...string) {
	if len(vals) == 0 {
		return
	}
	h[k] = vals
}

// Append adds the values to key k, not overwriting what was already stored at that key.
func (h Header) Append(k string, vals ...string) {
	if len(vals) == 0 {
		return
	}
	h[k] = append(h[k], vals...)
}

// Join joins any number of mds into a single Header.
// The order of values for each key is determined by the order in which
// the mds containing those values are presented to Join.
func Join(mds ...Header) Header {
	out := Header{}
	for _, h := range mds {
		for k, v := range h {
			out[k] = append(out[k], v...)
		}
	}
	return out
}
