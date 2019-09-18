package imagemanip

import (
	"fmt"
	"image"
	"reflect"
)

func Resize(imgSrc image.Image, bounds image.Rectangle) image.Image {
	var srcPix []uint8
	var dstPix []uint8
	var result image.Image
	switch imgSrc.(type) {
	case *image.RGBA:
		srcPix = imgSrc.(*image.RGBA).Pix
		r1 := image.NewRGBA(bounds)
		result = r1
		dstPix = r1.Pix
	case *image.NRGBA:
		srcPix = imgSrc.(*image.NRGBA).Pix
		r1 := image.NewNRGBA(bounds)
		result = r1
		dstPix = r1.Pix
	default:
		panic(fmt.Sprintf("Unsupported image type %v ", reflect.TypeOf(imgSrc).String()))
	}

	dstPos := 0
	for y := 0; y < result.Bounds().Dy(); y++ {
		for x := 0; x < result.Bounds().Dx(); x++ {
			srcY := y * imgSrc.Bounds().Dy() / result.Bounds().Dy()
			srcX := x * imgSrc.Bounds().Dx() / result.Bounds().Dx()
			srcPos := 4 * (srcX + imgSrc.Bounds().Dx()*srcY)
			dstPix[dstPos] = srcPix[srcPos]
			dstPix[dstPos+1] = srcPix[srcPos+1]
			dstPix[dstPos+2] = srcPix[srcPos+2]
			dstPix[dstPos+3] = srcPix[srcPos+3]
			dstPos += 4
		}
	}
	return result
}
