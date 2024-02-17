package mashup

import (
	"image"
	"sort"
)

func getProminentImageColors(image image.Image) (pixelMap map[pixelData][]*coordsData, keys []pixelData) { // map of RGBA:slice of all cords
	bounds := image.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	pixelMap = make(map[pixelData][]*coordsData)
	for y := 0; y < height; y++ { // loop through all pixels
		for x := 0; x < width; x++ {
			rgbaPixel := rgbaToPixel(image.At(x, y).RGBA())
			coords := pixelMap[rgbaPixel]
			pixelMap[rgbaPixel] = append(coords, &coordsData{X: x, Y: y})
		}
	}
	keys = mapKeys(pixelMap)
	sort.Slice(keys, func(i, j int) bool {
		return len(pixelMap[keys[i]]) > len(pixelMap[keys[j]])
	})
	return
}

func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) pixelData {
	return pixelData{uint8(r / 257), uint8(g / 257), uint8(b / 257), uint8(a / 257)}
}

func mapKeys(m map[pixelData][]*coordsData) []pixelData {
	s := make([]pixelData, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	return s
}

type pixelData struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

type coordsData struct {
	X int
	Y int
}
