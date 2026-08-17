package imageprocessor

import "github.com/h2non/bimg"

// BimgProcessor transforms images using the bimg binding for libvips.
type BimgProcessor struct{}

// NewBimgProcessor creates a libvips-backed image processor.
func NewBimgProcessor() *BimgProcessor {
	return &BimgProcessor{}
}

func (p *BimgProcessor) ProcessFile(inputPath string, variants []Variant) error {
	buffer, err := bimg.Read(inputPath)
	if err != nil {
		return err
	}
	return p.ProcessBuffer(buffer, variants)
}

func (p *BimgProcessor) ProcessBuffer(input []byte, variants []Variant) error {
	if err := validateVariants(variants); err != nil {
		return err
	}
	for _, variant := range variants {
		imageType, _ := bimgType(variant.Format)
		processed, err := bimg.NewImage(input).Process(bimg.Options{
			Width:   int(variant.Width),
			Height:  0,
			Quality: variant.Quality,
			Type:    imageType,
		})
		if err != nil {
			return err
		}
		if err := bimg.Write(variant.Path, processed); err != nil {
			return err
		}
	}
	return nil
}

func bimgType(format Format) (bimg.ImageType, error) {
	switch format {
	case FormatJPEG:
		return bimg.JPEG, nil
	case FormatWebP:
		return bimg.WEBP, nil
	default:
		return bimg.UNKNOWN, unsupportedFormatError(format)
	}
}

var _ Processor = (*BimgProcessor)(nil)
