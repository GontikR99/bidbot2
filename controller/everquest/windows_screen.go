package everquest

import (
	"errors"
	"fmt"
	imagemanip2 "github.com/gontikr99/bidbot2/controller/imagemanip"
	"github.com/lxn/win"
	"image"
	"strings"
	"syscall"
	"unsafe"
)

var (
	libUser32, _               = syscall.LoadLibrary("user32.dll")
	funcGetDesktopWindow, _    = syscall.GetProcAddress(syscall.Handle(libUser32), "GetDesktopWindow")
	funcEnumDisplayMonitors, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "EnumDisplayMonitors")
)

func enumerateWindows() (mapping map[win.HWND]string, err error) {
	mapping = make(map[win.HWND]string)
	ewProc := syscall.NewCallback(func(handle win.HWND, _ uintptr) uintptr {
		iwv, _, _ := isWindowVisible.Call(uintptr(handle))
		if iwv != 0 {
			length, _, _ := getWindowTextLength.Call(uintptr(handle))
			buffer := make([]uint16, length+1)
			rl, _, _ := getWindowText.Call(uintptr(handle), uintptr(unsafe.Pointer(&buffer[0])), uintptr(length+1))
			if rl != 0 {
				buffer = buffer[:rl]
				mapping[handle] = syscall.UTF16ToString(buffer)
			}
		}
		return 1
	})
	rv, _, errTmp := enumWindows.Call(ewProc, 0)
	if rv == 0 {
		err = errTmp
	}
	return
}

// Get the window handle for EverQuest
func findEverQuest() (win.HWND, error) {
	mapping, err := enumerateWindows()
	if err != nil {
		return 0, fmt.Errorf("Failed to enumerate windows: %v", err)
	}
	for handle, name := range mapping {
		if strings.Compare("EverQuest", name) == 0 {
			return handle, nil
		}
	}
	return 0, errors.New("No EverQuest window found")
}

// Bring the EverQuest window to the foreground
func raiseEverquest() error {
	handle, err := findEverQuest()
	if err != nil {
		return err
	}
	rv, _, tmpErr := setForegroundWindow.Call(uintptr(handle))
	if rv == 0 {
		return fmt.Errorf("Failed to set foreground window: %v", tmpErr)
	} else {
		return nil
	}
}

func getEqClientArea() (x int, y int, width int, height int, err error) {
	handle, err := findEverQuest()
	if err != nil {
		return
	}
	pt := point{0, 0}
	rv, _, tmpErr := screenToClient.Call(uintptr(handle), uintptr(unsafe.Pointer(&pt)))
	if rv == 0 {
		err = tmpErr
		return
	}
	x = -int(x)
	y = -int(y)
	wrect := rect{}
	rv, _, tmpErr = getClientRect.Call(uintptr(handle), uintptr(unsafe.Pointer(&wrect)))
	if rv == 0 {
		err = tmpErr
		return
	}
	width = int(wrect.right - wrect.left)
	height = int(wrect.bottom - wrect.top)
	return
}

// Capture the EverQuest client area as an image
func captureEverquest(bounds image.Rectangle) (img image.Image, err error) {
	x, y, width, height, err := getEqClientArea()
	if err != nil {
		return
	}
	err = raiseEverquest()
	if err != nil {
		return
	}
	if width > bounds.Dx() {
		width = bounds.Dx()
	}
	if height > bounds.Dy() {
		height = bounds.Dy()
	}
	img, err = captureImage(x+bounds.Min.X, y+bounds.Min.Y, width, height)
	return
}

func createImage(rect image.Rectangle) (img *image.RGBA, e error) {
	img = nil
	e = errors.New("Cannot create image.RGBA")

	defer func() {
		err := recover()
		if err == nil {
			e = nil
		}
	}()
	// image.NewRGBA may panic if rect is too large.
	img = image.NewRGBA(rect)

	return img, e
}

func drawText(theight int, text string) (image.Image, error) {
	width := 1024
	height := theight + 1
	rect := image.Rect(0, 0, width, height)
	img, err := createImage(rect)
	if err != nil {
		return nil, err
	}

	hwnd := getDesktopWindow()
	hdc := win.GetDC(hwnd)
	if hdc == 0 {
		return nil, errors.New("GetDC failed")
	}
	defer win.ReleaseDC(hwnd, hdc)

	memory_device := win.CreateCompatibleDC(hdc)
	if memory_device == 0 {
		return nil, errors.New("CreateCompatibleDC failed")
	}
	defer win.DeleteDC(memory_device)

	bitmap := win.CreateCompatibleBitmap(hdc, int32(width), int32(height))
	if bitmap == 0 {
		return nil, errors.New("CreateCompatibleBitmap failed")
	}
	defer win.DeleteObject(win.HGDIOBJ(bitmap))

	var header win.BITMAPINFOHEADER
	header.BiSize = uint32(unsafe.Sizeof(header))
	header.BiPlanes = 1
	header.BiBitCount = 32
	header.BiWidth = int32(width)
	header.BiHeight = int32(-height)
	header.BiCompression = win.BI_RGB
	header.BiSizeImage = 0

	// GetDIBits balks at using Go memory on some systems. The MSDN example uses
	// GlobalAlloc, so we'll do that too. See:
	// https://docs.microsoft.com/en-gb/windows/desktop/gdi/capturing-an-image
	bitmapDataSize := uintptr(((int64(width)*int64(header.BiBitCount) + 31) / 32) * 4 * int64(height))
	hmem := win.GlobalAlloc(win.GMEM_MOVEABLE, bitmapDataSize)
	defer win.GlobalFree(hmem)
	memptr := win.GlobalLock(hmem)
	defer win.GlobalUnlock(hmem)

	old := win.SelectObject(memory_device, win.HGDIOBJ(bitmap))
	if old == 0 {
		return nil, errors.New("SelectObject failed")
	}
	defer win.SelectObject(memory_device, old)

	fontParams := &win.LOGFONT{
		LfHeight:         int32(theight),
		LfWidth:          0,
		LfEscapement:     0,
		LfOrientation:    0,
		LfWeight:         win.FW_DONTCARE,
		LfItalic:         win.FALSE,
		LfUnderline:      win.FALSE,
		LfStrikeOut:      win.FALSE,
		LfCharSet:        win.DEFAULT_CHARSET,
		LfOutPrecision:   win.OUT_DEFAULT_PRECIS,
		LfClipPrecision:  win.CLIP_DEFAULT_PRECIS,
		LfQuality:        win.DEFAULT_QUALITY,
		LfPitchAndFamily: win.VARIABLE_PITCH,
	}
	arielU16 := syscall.StringToUTF16("ariel")
	copy(fontParams.LfFaceName[:], arielU16)
	font := win.CreateFontIndirect(fontParams)
	defer win.DeleteObject(win.HGDIOBJ(font))
	win.SetTextColor(memory_device, win.COLORREF(0xb08f6e))
	win.SetBkColor(memory_device, win.COLORREF(0))
	win.SelectObject(memory_device, win.HGDIOBJ(font))

	trect := win.RECT{0, 0, int32(width), int32(height)}
	win.SetRect(&trect, 0, 0, uint32(width), uint32(height))
	rv, _, err := drawTextW.Call(
		uintptr(memory_device),
		uintptr(unsafe.Pointer(&syscall.StringToUTF16(text)[0])),
		uintptr(len(syscall.StringToUTF16(text))-1),
		uintptr(unsafe.Pointer(&trect)),
		uintptr(win.DT_TOP|win.DT_LEFT|win.DT_NOCLIP|win.DT_NOPREFIX|win.DT_SINGLELINE),
	)
	if rv == 0 {
		return nil, err
	}

	if win.GetDIBits(hdc, bitmap, 0, uint32(height), (*uint8)(memptr), (*win.BITMAPINFO)(unsafe.Pointer(&header)), win.DIB_RGB_COLORS) == 0 {
		return nil, errors.New("GetDIBits failed")
	}
	i := 0
	src := uintptr(memptr)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			v0 := *(*uint8)(unsafe.Pointer(src))
			v1 := *(*uint8)(unsafe.Pointer(src + 1))
			v2 := *(*uint8)(unsafe.Pointer(src + 2))

			// BGRA => RGBA, and set A to 255
			img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = v2, v1, v0, 255

			i += 4
			src += 4
		}
	}

	return imagemanip2.Trim(img), nil
}

func captureImage(x, y, width, height int) (*image.RGBA, error) {
	rect := image.Rect(0, 0, width, height)
	img, err := createImage(rect)
	if err != nil {
		return nil, err
	}

	hwnd := getDesktopWindow()
	hdc := win.GetDC(hwnd)
	if hdc == 0 {
		return nil, errors.New("GetDC failed")
	}
	defer win.ReleaseDC(hwnd, hdc)

	memory_device := win.CreateCompatibleDC(hdc)
	if memory_device == 0 {
		return nil, errors.New("CreateCompatibleDC failed")
	}
	defer win.DeleteDC(memory_device)

	bitmap := win.CreateCompatibleBitmap(hdc, int32(width), int32(height))
	if bitmap == 0 {
		return nil, errors.New("CreateCompatibleBitmap failed")
	}
	defer win.DeleteObject(win.HGDIOBJ(bitmap))

	var header win.BITMAPINFOHEADER
	header.BiSize = uint32(unsafe.Sizeof(header))
	header.BiPlanes = 1
	header.BiBitCount = 32
	header.BiWidth = int32(width)
	header.BiHeight = int32(-height)
	header.BiCompression = win.BI_RGB
	header.BiSizeImage = 0

	// GetDIBits balks at using Go memory on some systems. The MSDN example uses
	// GlobalAlloc, so we'll do that too. See:
	// https://docs.microsoft.com/en-gb/windows/desktop/gdi/capturing-an-image
	bitmapDataSize := uintptr(((int64(width)*int64(header.BiBitCount) + 31) / 32) * 4 * int64(height))
	hmem := win.GlobalAlloc(win.GMEM_MOVEABLE, bitmapDataSize)
	defer win.GlobalFree(hmem)
	memptr := win.GlobalLock(hmem)
	defer win.GlobalUnlock(hmem)

	old := win.SelectObject(memory_device, win.HGDIOBJ(bitmap))
	if old == 0 {
		return nil, errors.New("SelectObject failed")
	}
	defer win.SelectObject(memory_device, old)

	if !win.BitBlt(memory_device, 0, 0, int32(width), int32(height), hdc, int32(x), int32(y), win.SRCCOPY) {
		return nil, errors.New("BitBlt failed")
	}

	if win.GetDIBits(hdc, bitmap, 0, uint32(height), (*uint8)(memptr), (*win.BITMAPINFO)(unsafe.Pointer(&header)), win.DIB_RGB_COLORS) == 0 {
		return nil, errors.New("GetDIBits failed")
	}

	i := 0
	src := uintptr(memptr)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			v0 := *(*uint8)(unsafe.Pointer(src))
			v1 := *(*uint8)(unsafe.Pointer(src + 1))
			v2 := *(*uint8)(unsafe.Pointer(src + 2))

			// BGRA => RGBA, and set A to 255
			img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = v2, v1, v0, 255

			i += 4
			src += 4
		}
	}

	return img, nil
}

func numActiveDisplays() int {
	var count int = 0
	enumDisplayMonitors(win.HDC(0), nil, syscall.NewCallback(countupMonitorCallback), uintptr(unsafe.Pointer(&count)))
	return count
}

func getDesktopWindow() win.HWND {
	ret, _, _ := syscall.Syscall(funcGetDesktopWindow, 0, 0, 0, 0)
	return win.HWND(ret)
}

func enumDisplayMonitors(hdc win.HDC, lprcClip *win.RECT, lpfnEnum uintptr, dwData uintptr) bool {
	ret, _, _ := syscall.Syscall6(funcEnumDisplayMonitors, 4,
		uintptr(hdc),
		uintptr(unsafe.Pointer(lprcClip)),
		lpfnEnum,
		dwData,
		0,
		0)
	return int(ret) != 0
}

func countupMonitorCallback(hMonitor win.HMONITOR, hdcMonitor win.HDC, lprcMonitor *win.RECT, dwData uintptr) uintptr {
	var count *int
	count = (*int)(unsafe.Pointer(dwData))
	*count = *count + 1
	return uintptr(1)
}
