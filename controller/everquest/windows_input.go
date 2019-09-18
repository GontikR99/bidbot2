package everquest

// Microsoft Windows platform specific functions.

import (
	"errors"
	"syscall"
	"time"
	"unsafe"
)

const (
	tapDelay        = 3 * time.Millisecond
	tapReleaseDelay = time.Millisecond
	slowTapDelay    = 25 * time.Millisecond
	clickDelay      = 100 * time.Millisecond
	vkShift         = 0x10
	vkControl       = 0x11
	vkMenu          = 0x12
)

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	sendInput           = user32.NewProc("SendInput")
	mapVirtualKeyExW    = user32.NewProc("MapVirtualKeyExW")
	enumWindows         = user32.NewProc("EnumWindows")
	getWindowText       = user32.NewProc("GetWindowTextW")
	getWindowTextLength = user32.NewProc("GetWindowTextLengthW")
	isWindowVisible     = user32.NewProc("IsWindowVisible")
	setForegroundWindow = user32.NewProc("SetForegroundWindow")
	getWindowRect       = user32.NewProc("GetWindowRect")
	getClientRect       = user32.NewProc("GetClientRect")
	screenToClient      = user32.NewProc("ScreenToClient")
	drawTextW           = user32.NewProc("DrawTextW")
)

type keyboardInput struct {
	wVk         uint16
	wScan       uint16
	dwFlags     uint32
	time        uint32
	dwExtraInfo uint64
}

type kbdinputUnion struct {
	inputType uint32
	ki        keyboardInput
	padding   uint64
}

type mouseInput struct {
	dx          int32
	dy          int32
	mouseData   uint32
	dwFlags     uint32
	time        uint32
	dwExtraInfo uint64
}

type mouseinputUnion struct {
	inputType uint32
	mi        mouseInput
}

type point struct {
	x int32
	y int32
}

type rect struct {
	left   int32
	top    int32
	right  int32
	bottom int32
}

// Typing text involves pressing appropriate keys on the keyboard.  For some keys (like capital letters),
// we need to press [Shift] first.  Keep a mapping from rune to what's needed to produce it
type codeMapping struct {
	keycode   uint16
	needShift bool
}

var keyMap = map[rune]codeMapping{
	'\n':     {13, false},
	'\b':     {0x08, false},
	'\u001b': {0x1b, false},
	'\t':     {0x09, false},
	' ':      {32, false},
	'0':      {48, false},
	'1':      {49, false},
	'2':      {50, false},
	'3':      {51, false},
	'4':      {52, false},
	'5':      {53, false},
	'6':      {54, false},
	'7':      {55, false},
	'8':      {56, false},
	'9':      {57, false},
	')':      {48, true},
	'!':      {49, true},
	'@':      {50, true},
	'#':      {51, true},
	'$':      {52, true},
	'%':      {53, true},
	'^':      {54, true},
	'&':      {55, true},
	'*':      {56, true},
	'(':      {57, true},
	'a':      {65, false},
	'b':      {66, false},
	'c':      {67, false},
	'd':      {68, false},
	'e':      {69, false},
	'f':      {70, false},
	'g':      {71, false},
	'h':      {72, false},
	'i':      {73, false},
	'j':      {74, false},
	'k':      {75, false},
	'l':      {76, false},
	'm':      {77, false},
	'n':      {78, false},
	'o':      {79, false},
	'p':      {80, false},
	'q':      {81, false},
	'r':      {82, false},
	's':      {83, false},
	't':      {84, false},
	'u':      {85, false},
	'v':      {86, false},
	'w':      {87, false},
	'x':      {88, false},
	'y':      {89, false},
	'z':      {90, false},
	'A':      {65, true},
	'B':      {66, true},
	'C':      {67, true},
	'D':      {68, true},
	'E':      {69, true},
	'F':      {70, true},
	'G':      {71, true},
	'H':      {72, true},
	'I':      {73, true},
	'J':      {74, true},
	'K':      {75, true},
	'L':      {76, true},
	'M':      {77, true},
	'N':      {78, true},
	'O':      {79, true},
	'P':      {80, true},
	'Q':      {81, true},
	'R':      {82, true},
	'S':      {83, true},
	'T':      {84, true},
	'U':      {85, true},
	'V':      {86, true},
	'W':      {87, true},
	'X':      {88, true},
	'Y':      {89, true},
	'Z':      {90, true},
	';':      {186, false},
	':':      {186, true},
	'=':      {187, false},
	'+':      {187, true},
	',':      {188, false},
	'<':      {188, true},
	'-':      {189, false},
	'_':      {189, true},
	'.':      {190, false},
	'>':      {190, true},
	'/':      {191, false},
	'?':      {191, true},
	'`':      {192, false},
	'~':      {192, true},
	'[':      {219, false},
	'{':      {219, true},
	'\\':     {220, false},
	'|':      {220, true},
	']':      {221, false},
	'}':      {221, true},
	'\'':     {222, false},
	'"':      {222, true},
}

func pressKeyInternal(keycode uint16, dwFlags uint32) error {
	scanCode, _, _ := mapVirtualKeyExW.Call(uintptr(keycode), 0, 0)
	var i kbdinputUnion
	i.inputType = 1 //INPUT_KEYBOARD
	i.ki.wVk = keycode
	i.ki.wScan = uint16(scanCode)
	i.ki.dwFlags = dwFlags
	ret, _, err := sendInput.Call(
		uintptr(1),
		uintptr(unsafe.Pointer(&i)),
		uintptr(unsafe.Sizeof(i)),
	)
	if ret == 0 {
		return err
	}
	return nil
}

// Depress the key with the specified key code
func pressKey(keycode uint16) error {
	return pressKeyInternal(keycode, 0)
}

// Release the key with the specified key code
func releaseKey(keycode uint16) error {
	return pressKeyInternal(keycode, 2 /* KEYEVENTF_KEYUP */)
}

// Press and then release keys causing the specified rune to be typed
func tap(character rune) error {
	cm, ok := keyMap[character]
	if !ok {
		return errors.New("Can't type character")
	}
	if cm.needShift {
		pressKey(vkShift)
		time.Sleep(tapDelay)
	}
	pressKey(cm.keycode)
	time.Sleep(tapDelay)
	releaseKey(cm.keycode)
	time.Sleep(tapReleaseDelay)
	if cm.needShift {
		releaseKey(vkShift)
		time.Sleep(tapReleaseDelay)
	}
	return nil
}

// Press and then release keys causing the specified rune to be typed.  Delay more than usual.
func tapSlow(character rune) error {
	cm, ok := keyMap[character]
	if !ok {
		return errors.New("Can't type character")
	}
	if cm.needShift {
		pressKey(vkShift)
		time.Sleep(slowTapDelay)
	}
	pressKey(cm.keycode)
	time.Sleep(slowTapDelay)
	releaseKey(cm.keycode)
	if cm.needShift {
		time.Sleep(slowTapDelay)
		releaseKey(vkShift)
	}
	return nil
}

// Press and release keys causing the specified string to be typed
func typewrite(text string) {
	for _, c := range text {
		tap(c)
		time.Sleep(tapDelay)
	}
}

func moveMouse(x int, y int) error {
	handle := getDesktopWindow()
	wrect := rect{}
	getClientRect.Call(uintptr(handle), uintptr(unsafe.Pointer(&wrect)))
	ipt := mouseinputUnion{
		inputType: 0, // INPUT_MOUSE
		mi: mouseInput{
			dx:      int32(65536 * x / int(wrect.right-wrect.left)),
			dy:      int32(65536 * y / int(wrect.bottom-wrect.top)),
			dwFlags: uint32(0x8000 /* | 0x4000 */ | 1), // MOUSEEVENTF_ABSOLUTE | MOUSEEVENTF_VIRTUALDESK | MOUSEEVENTF_MOVE
		},
	}
	ret, _, err := sendInput.Call(
		uintptr(1),
		uintptr(unsafe.Pointer(&ipt)),
		uintptr(unsafe.Sizeof(ipt)),
	)
	if ret == 0 {
		return err
	} else {
		return nil
	}
}

func clickInternal(flag uint32) error {
	ipt := mouseinputUnion{
		inputType: 0, // INPUT_MOUSE
		mi: mouseInput{
			dwFlags: flag,
		},
	}
	ret, _, err := sendInput.Call(
		uintptr(1),
		uintptr(unsafe.Pointer(&ipt)),
		uintptr(unsafe.Sizeof(ipt)),
	)
	if ret == 0 {
		return err
	} else {
		return nil
	}
}

func leftClick() error {
	time.Sleep(clickDelay)
	err := clickInternal(2) // MOUSEEVENTF_LEFTDOWN
	if err != nil {
		return err
	}
	time.Sleep(clickDelay)
	err = clickInternal(4) // MOUSEEVENTF_LEFTUP
	if err != nil {
		return err
	}
	return nil
}
