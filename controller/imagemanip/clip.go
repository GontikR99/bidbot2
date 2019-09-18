package imagemanip

import (
	"fmt"
	"image"
	"reflect"
)

func Clip(imgSrc image.Image, bounds image.Rectangle) image.Image {
	xmin := bounds.Min.X
	ymin := bounds.Min.Y
	width := bounds.Dx()
	height := bounds.Dy()

	if imgSrc.Bounds().Dx() < width {
		width = imgSrc.Bounds().Dx()
	}
	if imgSrc.Bounds().Dy() < height {
		height = imgSrc.Bounds().Dy()
	}

	var srcPix []uint8
	var dstPix []uint8
	var result image.Image
	switch imgSrc.(type) {
	case *image.RGBA:
		srcPix = imgSrc.(*image.RGBA).Pix
		r1 := image.NewRGBA(image.Rect(0, 0, width, height))
		result = r1
		dstPix = r1.Pix
	case *image.NRGBA:
		srcPix = imgSrc.(*image.NRGBA).Pix
		r1 := image.NewNRGBA(image.Rect(0, 0, width, height))
		result = r1
		dstPix = r1.Pix
	default:
		panic(fmt.Sprintf("Unsupported image type %v ", reflect.TypeOf(imgSrc).String()))
	}

	dstPos := 0
	for y := ymin; y < ymin+height; y++ {
		srcPos := 4*imgSrc.Bounds().Dx()*y + 4*xmin
		for x := xmin; x < xmin+width; x++ {
			dstPix[dstPos] = srcPix[srcPos]
			dstPix[dstPos+1] = srcPix[srcPos+1]
			dstPix[dstPos+2] = srcPix[srcPos+2]
			dstPix[dstPos+3] = srcPix[srcPos+3]
			srcPos += 4
			dstPos += 4
		}
	}
	return result
}
