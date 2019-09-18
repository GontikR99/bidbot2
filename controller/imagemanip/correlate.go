package imagemanip

import (
	"fmt"
	"image"
	"reflect"
	"runtime"
	"sort"
	"sync"
)

type MatchLocation struct {
	image.Rectangle
	Score float32
}

// Find the image 'needle' within 'haystack', returning a list of matches.  This
// matching binarizes the images using a threshold before matching.
func FindWithThreshold(haystack image.Image, needle image.Image) []MatchLocation {
	return correlate(haystack, needle, 0.85, threshold)
}

// Find the image `needle` within `haystack`, returning a list of matches.  This
// matching binarizes the images using a simple edge detection algorithm before matching.
func FindWithEdges(haystack image.Image, needle image.Image) []MatchLocation {
	return correlate(haystack, needle, 0.95, edgeDetect)
}

type byScoreDecreasing []MatchLocation

func (mls byScoreDecreasing) Len() int           { return len(mls) }
func (mls byScoreDecreasing) Less(i, j int) bool { return mls[i].Score > mls[j].Score }
func (mls byScoreDecreasing) Swap(i, j int)      { mls[i], mls[j] = mls[j], mls[i] }

// Use feature extraction and fast integer transform to locate `needle` within `haystack`
func correlate(haystack image.Image, needle image.Image, cutoff float32, features func(image.Image, int, int, uint16, uint16) (modulusImage, uint16)) []MatchLocation {
	var width, height int
	if haystack.Bounds().Dx() > needle.Bounds().Dx() {
		width = haystack.Bounds().Dx()
	} else {
		width = needle.Bounds().Dx()
	}
	if haystack.Bounds().Dy() > needle.Bounds().Dy() {
		height = haystack.Bounds().Dy()
	} else {
		height = needle.Bounds().Dy()
	}
	width = 1 << log2(width)
	height = 1 << log2(height)

	// extract features
	haystackFeatures, _ := features(haystack, width, height, 1, 0)
	needleFeatures, pixelSum := features(needle, width, height, uint16(modulus-1), 1)

	// reverse needle
	{
		begin := 0
		end := len(needleFeatures.pixels) - 1
		for begin < end {
			needleFeatures.pixels[begin], needleFeatures.pixels[end] = needleFeatures.pixels[end], needleFeatures.pixels[begin]
			begin++
			end--
		}
	}

	// Convolve (reversed) needle with haystack
	fit2(haystackFeatures)
	fit2(needleFeatures)

	for i := 0; i < len(haystackFeatures.pixels); i++ {
		needleFeatures.pixels[i] = uint16((uint64(needleFeatures.pixels[i]) * uint64(haystackFeatures.pixels[i])) % modulus)
	}
	iit2(needleFeatures)

	for i := 0; i < len(needleFeatures.pixels); i++ {
		needleFeatures.pixels[i] = uint16((uint64(needleFeatures.pixels[i]) + uint64(pixelSum)) % modulus)
	}

	// Select locations of best match
	bestLocsUnfiltered := make([]MatchLocation, 0)
	needleArea := float32(needle.Bounds().Dy() * needle.Bounds().Dx())
	for i := 0; i < len(needleFeatures.pixels); i++ {
		score := 1.0 - float32(needleFeatures.pixels[i])/needleArea
		if score > cutoff {
			x := i % width
			y := i / width
			x = (x + 1) % width
			y = (y + 1) % height
			if x+needle.Bounds().Dx() > haystack.Bounds().Dx() || y+needle.Bounds().Dy() > haystack.Bounds().Dy() {
				continue
			}
			bestLocsUnfiltered = append(bestLocsUnfiltered, MatchLocation{
				image.Rect(x, y, x+needle.Bounds().Dx(), y+needle.Bounds().Dy()),
				score})
		}
	}
	sort.Sort(byScoreDecreasing(bestLocsUnfiltered))

	type ulCorner struct {
		x int
		y int
	}

	// Filter out matches that are off by 1 or 2 pixels.
	seen := make(map[ulCorner]bool)
	bestLocs := make([]MatchLocation, 0)
	for _, bl := range bestLocsUnfiltered {
		var alreadySeen bool
		for dx := -2; dx <= 2; dx++ {
			for dy := -2; dy <= 2; dy++ {
				_, alreadySeen := seen[ulCorner{bl.Min.X + dx, bl.Min.Y + dy}]
				if alreadySeen {
					break
				}
			}
		}
		if !alreadySeen {
			seen[ulCorner{bl.Min.X, bl.Min.Y}] = true
			bestLocs = append(bestLocs, bl)
		}
	}

	return bestLocs
}

// Simple feature extraction algorithms
type modulusImage struct {
	pixels []uint16
	width  int
}

// Take a color image and turn it into a binary `modulusImage` by thresholding.
func threshold(imgSrc image.Image, width, height int, present, absent uint16) (modulusImage, uint16) {
	var pix []uint8
	switch imgSrc.(type) {
	case *image.RGBA:
		pix = imgSrc.(*image.RGBA).Pix
	case *image.NRGBA:
		pix = imgSrc.(*image.NRGBA).Pix
	default:
		panic(fmt.Sprintf("Unsupported image type %v ", reflect.TypeOf(imgSrc).String()))
	}
	resultBuf := make([]uint16, width*height)
	src, dst := 0, 0
	pixelCount := uint16(0)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if x+1 < imgSrc.Bounds().Dx() && y+1 < imgSrc.Bounds().Dy() {
				clvl := int32(pix[src]) + int32(pix[src+1]) + int32(pix[src+2])
				if clvl > 3*48 {
					resultBuf[dst] = present
					pixelCount += 1
				} else {
					resultBuf[dst] = absent
				}
				src += 4
			}
			dst += 1
		}
		src += 4
	}
	return modulusImage{resultBuf, width}, pixelCount
}

// Take a color image and turn it into a binary `modulusImage` by detecting edges.
func edgeDetect(imgSrc image.Image, width, height int, present, absent uint16) (modulusImage, uint16) {
	var pix []uint8
	switch imgSrc.(type) {
	case *image.RGBA:
		pix = imgSrc.(*image.RGBA).Pix
	case *image.NRGBA:
		pix = imgSrc.(*image.NRGBA).Pix
	default:
		panic(fmt.Sprintf("Unsupported image type %v ", reflect.TypeOf(imgSrc).String()))
	}
	resultBuf := make([]uint16, width*height)
	src, dst := 0, 0
	pixelCount := uint16(0)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if x+1 < imgSrc.Bounds().Dx() && y+1 < imgSrc.Bounds().Dy() {
				clvl := int32(pix[src]) + int32(pix[src+1]) + int32(pix[src+2])
				rlvl := int32(pix[src+4+0]) + int32(pix[src+4+1]) + int32(pix[src+4+2])
				delta := 4 * imgSrc.Bounds().Dx()
				dlvl := int32(pix[src+delta]) + int32(pix[src+delta+1]) + int32(pix[src+delta+2])
				if !(-132 < clvl-rlvl && clvl-rlvl < 132) || !(-132 < clvl-dlvl && clvl-dlvl < 132) {
					resultBuf[dst] = present
					pixelCount += 1
				} else {
					resultBuf[dst] = absent
				}
				src += 4
			}
			dst += 1
		}
		src += 4
	}
	return modulusImage{resultBuf, width}, pixelCount
}

// Field specification for performing number-theoretic transform.  We choose the prime field of size 61441, as it
// allows a root of unity that's a reasonably large power of two (4096), and fits within a uint16.
const (
	modulus   = uint64(61441) // field size (a prime number); group of units has size divisible by...
	rootOrder = 4096          // largest power of two dividing modulus-1; upper limit for FIT size.
	modRoot   = uint64(39003) // a primitive root of unity of degree `rootOrder` in that field
)

// `modRoot` ** `e` modulo `modulus`
func modPow(e uint64) uint64 {
	p := modRoot
	accum := uint64(1)
	for e != 0 {
		if e&1 == 1 {
			accum = (accum * p) % modulus
		}
		p = (p * p) % modulus
		e >>= 1
	}
	return accum
}

// inverse of `v` modulo `modulus`
func invert(v uint64) uint64 {
	s, old_s := int64(0), int64(1)
	r, old_r := int64(modulus), int64(v)

	for r != 0 {
		quotient := old_r / r
		old_r, r = r, old_r-quotient*r
		old_s, s = s, old_s-quotient*s
	}
	old_s = old_s % int64(modulus)
	if old_s < 0 {
		old_s += int64(modulus)
	}
	return uint64(old_s)
}

// in-place decimation-in-frequency fast integer transform, base 61441.  No bit reordering, since we're only
// using this to correlate.
func fit1(arr []uint16, offset int, stride int, logSize uint) {
	n := 1 << logSize
	if n > rootOrder {
		panic("Too big in fit1")
	}
	m := 1 << logSize
	for s := logSize; s >= 1; s-- {
		ws := modPow(rootOrder - rootOrder>>s)
		for k := 0; k < n; k += m {
			w := uint64(1)
			posLeft := offset + stride*k
			posRight := offset + stride*(k+m/2)
			for j := 0; j < m/2; j++ {
				u := uint64(arr[posLeft])
				t := uint64(arr[posRight])
				arr[posLeft] = uint16((u + t) % modulus)
				arr[posRight] = uint16(((modulus + u - t) * w) % modulus)
				w = (w * ws) % modulus
				posLeft += stride
				posRight += stride
			}
		}
		m >>= 1
	}
}

// in-place decimation-in-time fast inverse integer transform, base 61441.  No bit reordering, since we're only using this to
// correlate.
func iit1(arr []uint16, offset int, stride int, logSize uint) {
	n := 1 << logSize
	if n > rootOrder {
		panic("Too big in iit1")
	}
	m := 2
	posDiff := stride
	for s := uint(1); s <= logSize; s++ {
		wm := modPow(rootOrder >> s)
		posLeft := offset
		posRight := offset + posDiff
		for k := 0; k < n; k += m {
			w := uint64(1)
			for j := 0; j < m>>1; j++ {
				t := w * uint64(arr[posRight]) % modulus
				u := uint64(arr[posLeft])
				arr[posLeft] = uint16((u + t) % modulus)
				arr[posRight] = uint16((modulus + u - t) % modulus)
				w = (w * wm) % modulus
				posLeft += stride
				posRight += stride
			}
			posLeft += posDiff
			posRight += posDiff
		}
		m <<= 1
		posDiff <<= 1
	}
	inv := invert(1 << logSize)
	pos := offset
	for i := 0; i < n; i++ {
		arr[pos] = uint16((uint64(arr[pos]) * inv) % modulus)
		pos += stride
	}
}

func log2(n int) uint {
	ct := uint(0)
	value := n
	for value > 0 {
		ct++
		value >>= 1
	}
	if n == 1<<(ct-1) {
		return ct - 1
	} else {
		return ct
	}
}

// 2 dimensional parallel fast integer transform
func fit2(imageStr modulusImage) {
	imgData := imageStr.pixels
	width := imageStr.width
	height := len(imgData) / width
	logWidth := log2(width)
	logHeight := log2(height)
	if width != 1<<uint(logWidth) || height != 1<<uint(logHeight) {
		panic("Can only perform power-of-two FITs")
	}
	mp := runtime.GOMAXPROCS(0)

	wg := &sync.WaitGroup{}
	for i := 0; i < mp; i++ {
		wg.Add(1)
		go func(offset int) {
			for j := offset; j < height; j += mp {
				fit1(imgData, width*j, 1, logWidth)
				runtime.Gosched()
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	wg = &sync.WaitGroup{}
	for i := 0; i < mp; i++ {
		wg.Add(1)
		go func(offset int) {
			for j := offset; j < width; j += mp {
				fit1(imgData, j, width, logHeight)
				runtime.Gosched()
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

// 2 dimensional parallel fast inverse integer transform
func iit2(imageStr modulusImage) {
	imgData := imageStr.pixels
	width := imageStr.width
	height := len(imgData) / width
	logWidth := log2(width)
	logHeight := log2(height)
	if width != 1<<logWidth || height != 1<<logHeight {
		panic("Can only perform power-of-two FITs")
	}
	mp := runtime.GOMAXPROCS(0)

	wg := &sync.WaitGroup{}
	for i := 0; i < mp; i++ {
		wg.Add(1)
		go func(offset int) {
			for j := offset; j < width; j += mp {
				iit1(imgData, j, width, logHeight)
				runtime.Gosched()
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	wg = &sync.WaitGroup{}
	for i := 0; i < mp; i++ {
		wg.Add(1)
		go func(offset int) {
			for j := offset; j < height; j += mp {
				iit1(imgData, width*j, 1, logWidth)
				runtime.Gosched()
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}
