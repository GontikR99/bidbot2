package imagemanip

import (
	"fmt"
	"image"
	"reflect"
)

func findBounds(imgSrc image.Image) image.Rectangle {
	var pix []uint8
	switch imgSrc.(type) {
	case *image.RGBA:
		pix = imgSrc.(*image.RGBA).Pix
	case *image.NRGBA:
		pix = imgSrc.(*image.NRGBA).Pix
	default:
		panic(fmt.Sprintf("Unsupported image type %v ", reflect.TypeOf(imgSrc).String()))
	}
	minX, minY := imgSrc.Bounds().Max.X, imgSrc.Bounds().Max.Y
	maxX, maxY := 0, 0
	pxLoc := 0
	for y := 0; y < imgSrc.Bounds().Dy(); y++ {
		for x := 0; x < imgSrc.Bounds().Dx(); x++ {
			if pix[pxLoc] != 0 || pix[pxLoc+1] != 0 || pix[pxLoc+2] != 0 {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
			pxLoc += 4
		}
	}
	if minX > maxX || minY > maxY {
		return imgSrc.Bounds()
	} else {
		return image.Rect(minX, minY, maxX, maxY)
	}
}

// Cut off all columns of black pixels on the right-hand side of the image
func Trim(imgSrc image.Image) image.Image {
	bounds := findBounds(imgSrc)
	bounds = image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Max.X+1, bounds.Max.Y+1)
	return Clip(imgSrc, bounds)
}
