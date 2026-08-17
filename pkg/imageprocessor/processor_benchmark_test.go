package imageprocessor

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"path/filepath"
	"testing"
)

func BenchmarkProcessorVariants(b *testing.B) {
	input := benchmarkJPEG(b, 2400, 1600)

	benchmarks := []struct {
		name      string
		processor func() (Processor, error)
	}{
		{
			name: "bimg",
			processor: func() (Processor, error) {
				return NewBimgProcessor(), nil
			},
		},
		{
			name: "govips",
			processor: func() (Processor, error) {
				return NewGovipsProcessor()
			},
		},
	}

	for _, benchmark := range benchmarks {
		b.Run(benchmark.name, func(b *testing.B) {
			processor, err := benchmark.processor()
			if err != nil {
				b.Fatal(err)
			}
			dir := b.TempDir()
			variants := []Variant{
				{Path: filepath.Join(dir, "thumbnail.webp"), Width: 480, Quality: 90, Format: FormatWebP},
				{Path: filepath.Join(dir, "full.webp"), Width: 1920, Quality: 90, Format: FormatWebP},
			}

			b.SetBytes(int64(len(input)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := processor.ProcessBuffer(input, variants); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func benchmarkJPEG(b *testing.B, width, height int) []byte {
	b.Helper()
	source := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			source.SetNRGBA(x, y, color.NRGBA{
				R: uint8((x + y) % 256),
				G: uint8((x*3 + y) % 256),
				B: uint8((x + y*3) % 256),
				A: 255,
			})
		}
	}

	var output bytes.Buffer
	if err := jpeg.Encode(&output, source, &jpeg.Options{Quality: 92}); err != nil {
		b.Fatal(err)
	}
	return output.Bytes()
}
