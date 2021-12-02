package cache

type Option func(*options)

type options struct {
	root   string
	format FormatType
}

func WithLocalDir(root string) Option {
	return func(opts *options) {
		opts.root = root
	}
}

func WithFormat(formatType FormatType) Option {
	return func(opts *options) {
		opts.format = formatType
	}
}
