package balancer

type CustomOption struct {
	DC   string
	Tags []string
}

type Option func(*CustomOption)

func WithDC(dc string) Option {
	return func(o *CustomOption) {
		o.DC = dc
	}
}

func WithTags(tags ...string) Option {
	return func(o *CustomOption) {
		o.Tags = tags
	}
}
