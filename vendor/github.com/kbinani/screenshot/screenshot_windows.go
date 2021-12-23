package screenshot

import (
	"errors"
	"github.com/kbinani/screenshot/internal/util"
	win "github.com/lxn/win"
	"image"
	"runtime"
	"syscall"
	"unsafe"
)

var (
	libUser32, _               = syscall.LoadLibrary("user32.dll")
	funcGetDesktopWindow, _    = syscall.GetProcAddress(syscall.Handle(libUser32), "GetDesktopWindow")
	funcEnumDisplayMonitors, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "EnumDisplayMonitors")
	funcGetMonitorInfo, _      = syscall.GetProcAddress(syscall.Handle(libUser32), "GetMonitorInfoW")
	funcEnumDisplaySettings, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "EnumDisplaySettingsW")
)

type Capturer struct {
	width  int
	height int
	x      int
	y      int
	hwnd   win.HWND
	hdc    win.HDC
	memdev win.HDC
	bitmap win.HBITMAP
	memptr unsafe.Pointer
	hmem   win.HGLOBAL
	header win.BITMAPINFOHEADER
}

func freeCapturer(c *Capturer) {
	defer win.ReleaseDC(c.hwnd, c.hdc)
	defer win.DeleteDC(c.memdev)
	defer win.DeleteObject(win.HGDIOBJ(c.bitmap))
	defer win.GlobalFree(c.hmem)
	defer win.GlobalUnlock(c.hmem)
}

func (c *Capturer) NextFrame() ([]byte, error) {
	old := win.SelectObject(c.memdev, win.HGDIOBJ(c.bitmap))
	if old == 0 {
		return nil, errors.New("SelectObject failed")
	}
	defer win.SelectObject(c.memdev, old)

	if !win.BitBlt(c.memdev, 0, 0, int32(c.width), int32(c.height), c.hdc, int32(c.x), int32(c.y), win.SRCCOPY) {
		return nil, errors.New("BitBlt failed")
	}

	if win.GetDIBits(
		c.hdc, c.bitmap, 0, uint32(c.height), (*uint8)(c.memptr), (*win.BITMAPINFO)(unsafe.Pointer(&c.header)),
		win.DIB_RGB_COLORS,
	) == 0 {
		return nil, errors.New("GetDIBits failed")
	}

	i := 0
	src := uintptr(c.memptr)
	dst := make([]byte, c.width*c.height*4)
	for y := 0; y < c.height; y++ {
		for x := 0; x < c.width; x++ {
			v0 := *(*uint8)(unsafe.Pointer(src))
			v1 := *(*uint8)(unsafe.Pointer(src + 1))
			v2 := *(*uint8)(unsafe.Pointer(src + 2))

			dst[i] = v2
			dst[i+1] = v1
			dst[i+2] = v0
			dst[i+3] = 255

			// BGRA => RGBA, and set A to 255
			//img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = v2, v1, v0, 255

			i += 4
			src += 4
		}
	}

	return dst, nil
}

func NewCapturer(x, y, width, height int) (*Capturer, error) {
	c := &Capturer{
		width:  width,
		height: height,
		x:      x,
		y:      y,
	}
	runtime.SetFinalizer(c, freeCapturer)

	c.hwnd = getDesktopWindow()
	c.hdc = win.GetDC(c.hwnd)
	if c.hdc == 0 {
		return nil, errors.New("GetDC failed")
	}

	c.memdev = win.CreateCompatibleDC(c.hdc)
	if c.memdev == 0 {
		return nil, errors.New("CreateCompatibleDC failed")
	}

	c.bitmap = win.CreateCompatibleBitmap(c.hdc, int32(width), int32(height))
	if c.bitmap == 0 {
		return nil, errors.New("CreateCompatibleBitmap failed")
	}

	c.header = win.BITMAPINFOHEADER{}
	c.header.BiSize = uint32(unsafe.Sizeof(c.header))
	c.header.BiPlanes = 1
	c.header.BiBitCount = 32
	c.header.BiWidth = int32(width)
	c.header.BiHeight = int32(-height)
	c.header.BiCompression = win.BI_RGB
	c.header.BiSizeImage = 0

	// GetDIBits balks at using Go memory on some systems. The MSDN example uses
	// GlobalAlloc, so we'll do that too. See:
	// https://docs.microsoft.com/en-gb/windows/desktop/gdi/capturing-an-image
	bitmapDataSize := uintptr(((int64(width)*int64(c.header.BiBitCount) + 31) / 32) * 4 * int64(height))
	c.hmem = win.GlobalAlloc(win.GMEM_MOVEABLE, bitmapDataSize)
	c.memptr = win.GlobalLock(c.hmem)

	return c, nil
}

func Capture(x, y, width, height int) (*image.RGBA, error) {
	rect := image.Rect(0, 0, width, height)
	img, err := util.CreateImage(rect)
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

	if win.GetDIBits(
		hdc, bitmap, 0, uint32(height), (*uint8)(memptr), (*win.BITMAPINFO)(unsafe.Pointer(&header)), win.DIB_RGB_COLORS,
	) == 0 {
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

func NumActiveDisplays() int {
	var count int = 0
	enumDisplayMonitors(win.HDC(0), nil, syscall.NewCallback(countupMonitorCallback), uintptr(unsafe.Pointer(&count)))
	return count
}

func GetDisplayBounds(displayIndex int) image.Rectangle {
	var ctx getMonitorBoundsContext
	ctx.Index = displayIndex
	ctx.Count = 0
	enumDisplayMonitors(win.HDC(0), nil, syscall.NewCallback(getMonitorBoundsCallback), uintptr(unsafe.Pointer(&ctx)))
	return image.Rect(
		int(ctx.Rect.Left), int(ctx.Rect.Top),
		int(ctx.Rect.Right), int(ctx.Rect.Bottom),
	)
}

func getDesktopWindow() win.HWND {
	ret, _, _ := syscall.Syscall(funcGetDesktopWindow, 0, 0, 0, 0)
	return win.HWND(ret)
}

func enumDisplayMonitors(hdc win.HDC, lprcClip *win.RECT, lpfnEnum uintptr, dwData uintptr) bool {
	ret, _, _ := syscall.Syscall6(
		funcEnumDisplayMonitors, 4,
		uintptr(hdc),
		uintptr(unsafe.Pointer(lprcClip)),
		lpfnEnum,
		dwData,
		0,
		0,
	)
	return int(ret) != 0
}

func countupMonitorCallback(hMonitor win.HMONITOR, hdcMonitor win.HDC, lprcMonitor *win.RECT, dwData uintptr) uintptr {
	var count *int
	count = (*int)(unsafe.Pointer(dwData))
	*count = *count + 1
	return uintptr(1)
}

type getMonitorBoundsContext struct {
	Index int
	Rect  win.RECT
	Count int
}

func getMonitorBoundsCallback(
	hMonitor win.HMONITOR, hdcMonitor win.HDC, lprcMonitor *win.RECT, dwData uintptr,
) uintptr {
	var ctx *getMonitorBoundsContext
	ctx = (*getMonitorBoundsContext)(unsafe.Pointer(dwData))
	if ctx.Count != ctx.Index {
		ctx.Count = ctx.Count + 1
		return uintptr(1)
	}

	if realSize := getMonitorRealSize(hMonitor); realSize != nil {
		ctx.Rect = *realSize
	} else {
		ctx.Rect = *lprcMonitor
	}

	return uintptr(0)
}

type _MONITORINFOEX struct {
	win.MONITORINFO
	DeviceName [win.CCHDEVICENAME]uint16
}

const _ENUM_CURRENT_SETTINGS = 0xFFFFFFFF

type _DEVMODE struct {
	_            [68]byte
	DmSize       uint16
	_            [6]byte
	DmPosition   win.POINT
	_            [86]byte
	DmPelsWidth  uint32
	DmPelsHeight uint32
	_            [40]byte
}

// getMonitorRealSize makes a call to GetMonitorInfo
// to obtain the device name for the monitor handle
// provided to the method.
//
// With the device name, EnumDisplaySettings is called to
// obtain the current configuration for the monitor, this
// information includes the real resolution of the monitor
// rather than the scaled version based on DPI.
//
// If either handle calls fail, it will return a nil
// allowing the calling method to use the bounds information
// returned by EnumDisplayMonitors which may be affected
// by DPI.
func getMonitorRealSize(hMonitor win.HMONITOR) *win.RECT {
	info := _MONITORINFOEX{}
	info.CbSize = uint32(unsafe.Sizeof(info))

	ret, _, _ := syscall.Syscall(funcGetMonitorInfo, 2, uintptr(hMonitor), uintptr(unsafe.Pointer(&info)), 0)
	if ret == 0 {
		return nil
	}

	devMode := _DEVMODE{}
	devMode.DmSize = uint16(unsafe.Sizeof(devMode))

	if ret, _, _ := syscall.Syscall(
		funcEnumDisplaySettings, 3, uintptr(unsafe.Pointer(&info.DeviceName[0])), _ENUM_CURRENT_SETTINGS,
		uintptr(unsafe.Pointer(&devMode)),
	); ret == 0 {
		return nil
	}

	return &win.RECT{
		Left:   devMode.DmPosition.X,
		Right:  devMode.DmPosition.X + int32(devMode.DmPelsWidth),
		Top:    devMode.DmPosition.Y,
		Bottom: devMode.DmPosition.Y + int32(devMode.DmPelsHeight),
	}
}
