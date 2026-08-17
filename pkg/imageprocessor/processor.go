// Package imageprocessor defines the image transformation contract used by GOAPI.
// Concrete engines stay behind Processor so callers do not depend on a binding API.
package imageprocessor

// Format identifies an encoded output format.
type Format uint8

const (
	FormatUnknown Format = iota
	FormatJPEG
	FormatWebP
)

// Variant describes one image derived from a shared input.
type Variant struct {
	Path    string
	Width   uint
	Quality int
	Format  Format
}

// Processor creates one or more encoded variants from an image.
type Processor interface {
	ProcessFile(inputPath string, variants []Variant) error
	ProcessBuffer(input []byte, variants []Variant) error
}
