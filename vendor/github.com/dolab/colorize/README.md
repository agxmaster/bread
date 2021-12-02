# colorize

Simple golang command line colorize api for outputs highlight.

# Install

```go
go get -u github.com/dolab/colorize
```

# Usage

- Standard

```go
import "github.com/dolab/colorize"

func main() {
    brush := colorize.New("yellow+bBuih:black+h")
    brush.Paint("Hello, colorize!")
}
```

- Customize

```go
import "github.com/dolab/colorize"

func main() {
    brush := colorize.New("yellow+bBuih:black+h")

    // gain colors
    colorDraw, colorClean := brush.Colour()

    // custom output
    buf := bytes.NewBufferString(colorDraw)
    buf.WriteString("Hello, colorize!")
    buf.WriteString(colorClean)

    buf.WriteTo(os.Stdout)
}
```


# Format

> `foreground_color`+`color_attributes`:`background_color`+`color_attributes`

## buildin colors

- ColorRed
- ColorGreen
- ColorYellow
- ColorBlue
- ColorMagenta
- ColorCyan
- ColorWhite
- ColorGray
- ColorBlack

## available attributes
- b = bold foreground
- B = blink foreground
- u = underline foreground
- i = inverse
- h = high intensity (bright) foreground, background

# Windows support?

see [go-colorable](https://github.com/mattn/go-colorable)

# License

MIT

# Author

Spring MC
