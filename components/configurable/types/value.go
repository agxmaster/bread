package types

type ValueItem struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Offline bool   `json:"offline"`
}

type OutputItem struct {
	Value string `json:"value"`
	Name  string `json:"name"`
}

type Output struct {
	Available   []OutputItem        `json:"available"`
	All         []OutputItem        `json:"-"`
	Default     []OutputItem        `json:"default"`
	Required    bool                `json:"required"`
	Keys        map[string]struct{} `json:"-"`
	DefaultKeys map[string]struct{} `json:"-"`
}
