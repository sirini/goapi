package imageprocessor

import (
	"fmt"
	"os"
	"sync"

	"github.com/davidbyttow/govips/v2/vips"
)

// GovipsProcessor transforms images using the govips binding for libvips.
type GovipsProcessor struct{}

var (
	govipsStartupOnce sync.Once
	govipsStartupErr  error
)

// NewGovipsProcessor initializes govips once for the process and creates a processor.
// libvips intentionally remains running for the lifetime of the server.
func NewGovipsProcessor() (*GovipsProcessor, error) {
	govipsStartupOnce.Do(startGovips)
	if govipsStartupErr != nil {
		return nil, fmt.Errorf("start govips: %w", govipsStartupErr)
	}
	return &GovipsProcessor{}, nil
}

func startGovips() {
	defer func() {
		if recovered := recover(); recovered != nil {
			govipsStartupErr = fmt.Errorf("%v", recovered)
		}
	}()
	vips.LoggingSettings(nil, vips.LogLevelWarning)
	vips.Startup(nil)
}

func (p *GovipsProcessor) ProcessFile(inputPath string, variants []Variant) error {
	if err := validateVariants(variants); err != nil {
		return err
	}
	if len(variants) == 0 {
		return nil
	}

	image, err := vips.NewImageFromFile(inputPath)
	if err != nil {
		return fmt.Errorf("load image file: %w", err)
	}
	defer image.Close()

	return p.process(image, variants)
}

func (p *GovipsProcessor) ProcessBuffer(input []byte, variants []Variant) error {
	if err := validateVariants(variants); err != nil {
		return err
	}
	if len(variants) == 0 {
		return nil
	}

	image, err := vips.NewImageFromBuffer(input)
	if err != nil {
		return fmt.Errorf("load image buffer: %w", err)
	}
	defer image.Close()

	return p.process(image, variants)
}

func (p *GovipsProcessor) process(source *vips.ImageRef, variants []Variant) error {
	if err := source.AutoRotate(); err != nil {
		return fmt.Errorf("auto-rotate image: %w", err)
	}

	for _, variant := range variants {
		image, err := source.Copy()
		if err != nil {
			return fmt.Errorf("copy image for %q: %w", variant.Path, err)
		}

		if variant.Width > 0 {
			scale := float64(variant.Width) / float64(image.Width())
			if err := image.Resize(scale, vips.KernelLanczos3); err != nil {
				image.Close()
				return fmt.Errorf("resize image for %q: %w", variant.Path, err)
			}
		}

		output, err := exportGovipsImage(image, variant)
		image.Close()
		if err != nil {
			return fmt.Errorf("encode image for %q: %w", variant.Path, err)
		}
		if err := os.WriteFile(variant.Path, output, 0o644); err != nil {
			return fmt.Errorf("write image %q: %w", variant.Path, err)
		}
	}
	return nil
}

func exportGovipsImage(image *vips.ImageRef, variant Variant) ([]byte, error) {
	switch variant.Format {
	case FormatJPEG:
		params := vips.NewJpegExportParams()
		params.Quality = variant.Quality
		output, _, err := image.ExportJpeg(params)
		return output, err
	case FormatWebP:
		params := vips.NewWebpExportParams()
		params.Quality = variant.Quality
		output, _, err := image.ExportWebp(params)
		return output, err
	default:
		return nil, unsupportedFormatError(variant.Format)
	}
}

func validateVariants(variants []Variant) error {
	for _, variant := range variants {
		switch variant.Format {
		case FormatJPEG, FormatWebP:
		default:
			return unsupportedFormatError(variant.Format)
		}
	}
	return nil
}

func unsupportedFormatError(format Format) error {
	return fmt.Errorf("unsupported image format: %d", format)
}

var _ Processor = (*GovipsProcessor)(nil)
