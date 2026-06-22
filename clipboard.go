package main

import (
	"crypto/md5"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"sync"
	"time"
	"unsafe"

	"syscall"
)

var (
	user32          = syscall.NewLazyDLL("user32.dll")
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	openClipboard   = user32.NewProc("OpenClipboard")
	closeClipboard  = user32.NewProc("CloseClipboard")
	getClipboardData = user32.NewProc("GetClipboardData")
	isClipboardFormatAvailable = user32.NewProc("IsClipboardFormatAvailable")

	globalLock   = kernel32.NewProc("GlobalLock")
	globalUnlock = kernel32.NewProc("GlobalUnlock")
	globalSize   = kernel32.NewProc("GlobalSize")

	CF_DIB  = 8
	CF_TEXT = 1
)

type BITMAPINFOHEADER struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

type ClipboardMonitor struct {
	lastHash  [16]byte
	hasHash   bool
	mu        sync.Mutex
	tempDir   string
}

func NewClipboardMonitor() (*ClipboardMonitor, error) {
	dir := filepath.Join(os.TempDir(), "cc-clip-monitor")
	os.MkdirAll(dir, 0755)
	return &ClipboardMonitor{tempDir: dir}, nil
}

func (cm *ClipboardMonitor) CheckNewImage() (string, bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	r1, _, _ := isClipboardFormatAvailable.Call(uintptr(CF_DIB))
	if r1 == 0 {
		return "", false
	}

	r1, _, _ = openClipboard.Call(0)
	if r1 == 0 {
		return "", false
	}
	defer closeClipboard.Call()

	h, _, _ := getClipboardData.Call(uintptr(CF_DIB))
	if h == 0 {
		return "", false
	}

	ptr, _, _ := globalLock.Call(h)
	if ptr == 0 {
		return "", false
	}
	size, _, _ := globalSize.Call(h)

	data := make([]byte, size)
	copy(data, unsafe.Slice((*byte)(unsafe.Pointer(ptr)), size))
	globalUnlock.Call(ptr)

	if len(data) < 40 {
		return "", false
	}

	header := (*BITMAPINFOHEADER)(unsafe.Pointer(&data[0]))
	if header.BitCount != 32 && header.BitCount != 24 {
		return "", false
	}

	img := dibToImage(header, data)
	if img == nil {
		return "", false
	}

	hash := md5.Sum(data)
	if cm.hasHash && hash == cm.lastHash {
		return "", false
	}
	cm.lastHash = hash
	cm.hasHash = true

	path := filepath.Join(cm.tempDir, fmt.Sprintf("clip-%d.png", time.Now().UnixMilli()))
	f, err := os.Create(path)
	if err != nil {
		return "", false
	}
	defer f.Close()
	png.Encode(f, img)

	return path, true
}

func dibToImage(header *BITMAPINFOHEADER, data []byte) image.Image {
	w := int(header.Width)
	h := int(header.Height)
	if h < 0 {
		h = -h
	}
	bpp := int(header.BitCount) / 8
	rowSize := (w*bpp + 3) &^ 3
	pixelOffset := int(header.Size)

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		srcY := h - 1 - y
		if header.Height < 0 {
			srcY = y
		}
		srcOff := pixelOffset + srcY*rowSize
		for x := 0; x < w; x++ {
			srcIdx := srcOff + x*bpp
			if srcIdx+bpp > len(data) {
				continue
			}
			dstIdx := (y*w + x) * 4
			if bpp >= 4 {
				img.Pix[dstIdx+0] = data[srcIdx+2]
				img.Pix[dstIdx+1] = data[srcIdx+1]
				img.Pix[dstIdx+2] = data[srcIdx+0]
				img.Pix[dstIdx+3] = data[srcIdx+3]
			} else {
				img.Pix[dstIdx+0] = data[srcIdx+2]
				img.Pix[dstIdx+1] = data[srcIdx+1]
				img.Pix[dstIdx+2] = data[srcIdx+0]
				img.Pix[dstIdx+3] = 255
			}
		}
	}
	return img
}

func (cm *ClipboardMonitor) SetText(text string) error {
	return writeClipboardText(text)
}
