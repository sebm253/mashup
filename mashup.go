package mashup

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
)

// JPEG

// NewJPEGInput creates a new Input for JPEGs
func NewJPEGInput(r io.Reader) *Input {
	return &Input{
		in:               r,
		name:             "jpeg",
		magic:            "jpg",
		decodeFunc:       jpeg.Decode,
		decodeConfigFunc: jpeg.DecodeConfig,
	}
}

// NewJPEGOutput creates a new Output for JPEGs with the given quality
func NewJPEGOutput(w io.Writer, quality int) *Output {
	return &Output{
		w:          w,
		encodeFunc: jpegEncodeFunc(quality),
	}
}

// PNG

// NewPNGInput creates a new Input for PNGs
func NewPNGInput(r io.Reader) *Input {
	return &Input{
		in:               r,
		name:             "png",
		magic:            "png",
		decodeFunc:       png.Decode,
		decodeConfigFunc: png.DecodeConfig,
	}
}

// NewPNGOutput creates a new Output for PNGs
func NewPNGOutput(w io.Writer) *Output {
	return &Output{
		w:          w,
		encodeFunc: png.Encode,
	}
}

// custom

// NewCustomInput creates a new Input for types not supported by this library
func NewCustomInput(r io.Reader, name, magic string, decodeFunc DecodeFunc, decodeConfigFunc DecodeConfigFunc) *Input {
	return &Input{
		in:               r,
		name:             name,
		magic:            magic,
		decodeFunc:       decodeFunc,
		decodeConfigFunc: decodeConfigFunc,
	}
}

// NewCustomOutput creates a new Output for types not supported by this library
func NewCustomOutput(w io.Writer, encodeFunc EncodeFunc) *Output {
	return &Output{
		w:          w,
		encodeFunc: encodeFunc,
	}
}

// magic

type Input struct {
	in io.Reader

	name  string
	magic string

	decodeFunc       DecodeFunc
	decodeConfigFunc DecodeConfigFunc
}

type Output struct {
	w io.Writer

	encodeFunc EncodeFunc
}

// Mashup creates a color mashup of src and dst by computing and replacing their most prominent colors.
//
// Specify the maximum amount of colors that should be replaced using the maxColors parameter.
// If the amount of colors in either src or dst is less than the maxColors value, this amount will become the maximum.
//
// Additionally, if the src and dst are not of the same type, both formats are registered.
func Mashup(src, dst *Input, out *Output, maxColors int) error {
	image.RegisterFormat(src.name, src.magic, src.decodeFunc, src.decodeConfigFunc)

	if dst.name != src.name || dst.magic != src.magic {
		image.RegisterFormat(dst.name, dst.magic, dst.decodeFunc, dst.decodeConfigFunc)
	}

	srcImage, _, err := image.Decode(src.in)
	if err != nil {
		return fmt.Errorf("mashup: could not decode src to image: %w", err)
	}
	_, srcSortedKeys := getProminentImageColors(srcImage)

	dstImage, _, err := image.Decode(dst.in)
	if err != nil {
		return fmt.Errorf("mashup: could not decode dst to image: %w", err)
	}
	dstPixels, dstSortedKeys := getProminentImageColors(dstImage)

	amount := min(maxColors, len(dstSortedKeys), len(srcSortedKeys))
	dstSortedKeys = dstSortedKeys[:amount]

	modifiedImage := image.NewRGBA(dstImage.Bounds())
	draw.Draw(modifiedImage, dstImage.Bounds(), dstImage, image.Point{}, draw.Over)

	for i, key := range dstSortedKeys {
		srcColor := srcSortedKeys[i]
		for _, coords := range dstPixels[key] {
			modifiedImage.Set(coords.X, coords.Y, color.RGBA{R: srcColor.R, G: srcColor.G, B: srcColor.B, A: srcColor.A})
		}
	}
	if err := out.encodeFunc(out.w, modifiedImage); err != nil {
		return fmt.Errorf("mashup: could not encode image: %w", err)
	}
	return nil
}

// helpers

type DecodeFunc func(io.Reader) (image.Image, error)
type DecodeConfigFunc func(io.Reader) (image.Config, error)

type EncodeFunc func(io.Writer, image.Image) error

func jpegEncodeFunc(quality int) EncodeFunc {
	return func(w io.Writer, img image.Image) error {
		return jpeg.Encode(w, img, &jpeg.Options{
			Quality: quality,
		})
	}
}
