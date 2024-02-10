package mashup

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/pkg/errors"
)

// JPEG

func NewJPEGInput(r io.Reader) *Input {
	return &Input{
		in:               r,
		name:             "jpeg",
		magic:            "jpg",
		decodeFunc:       jpeg.Decode,
		decodeConfigFunc: jpeg.DecodeConfig,
	}
}

func NewJPEGOutput(w io.Writer, quality int) *Output {
	return &Output{
		w:          w,
		encodeFunc: jpegEncodeFunc(quality),
	}
}

// PNG

func NewPNGInput(r io.Reader) *Input {
	return &Input{
		in:               r,
		name:             "png",
		magic:            "png",
		decodeFunc:       png.Decode,
		decodeConfigFunc: png.DecodeConfig,
	}
}

func NewPNGOutput(w io.Writer) *Output {
	return &Output{
		w:          w,
		encodeFunc: png.Encode,
	}
}

// magic

type Input struct {
	in io.Reader

	name  string
	magic string

	decodeFunc       func(io.Reader) (image.Image, error)
	decodeConfigFunc func(io.Reader) (image.Config, error)
}

type Output struct {
	w io.Writer

	encodeFunc func(io.Writer, image.Image) error
}

func Mashup(src, dst Input, out Output, maxColors int) error {
	image.RegisterFormat(src.name, src.magic, src.decodeFunc, src.decodeConfigFunc)

	if dst.name != src.name || dst.magic != src.magic {
		image.RegisterFormat(dst.name, dst.magic, dst.decodeFunc, dst.decodeConfigFunc)
	}

	srcImage, _, err := image.Decode(src.in)
	if err != nil {
		return errors.Wrap(err, "mashup: could not decode source to image")
	}
	_, srcSortedKeys := getProminentImageColors(srcImage)

	dstImage, _, err := image.Decode(dst.in)
	if err != nil {
		return errors.Wrap(err, "mashup: could not decode dst to image")
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
		return errors.Wrap(err, "mashup: could not encode image")
	}
	return nil
}

func jpegEncodeFunc(quality int) func(io.Writer, image.Image) error {
	return func(w io.Writer, img image.Image) error {
		return jpeg.Encode(w, img, &jpeg.Options{
			Quality: quality,
		})
	}
}
