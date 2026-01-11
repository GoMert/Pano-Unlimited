//go:build windows
// +build windows

package clipboard

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                     = windows.NewLazySystemDLL("user32.dll")
	kernel32                   = windows.NewLazySystemDLL("kernel32.dll")
	openClipboard              = user32.NewProc("OpenClipboard")
	closeClipboard             = user32.NewProc("CloseClipboard")
	emptyClipboard             = user32.NewProc("EmptyClipboard")
	getClipboardData           = user32.NewProc("GetClipboardData")
	setClipboardData           = user32.NewProc("SetClipboardData")
	isClipboardFormatAvailable = user32.NewProc("IsClipboardFormatAvailable")
	globalLock                 = kernel32.NewProc("GlobalLock")
	globalUnlock               = kernel32.NewProc("GlobalUnlock")
	globalSize                 = kernel32.NewProc("GlobalSize")
	globalAlloc                = kernel32.NewProc("GlobalAlloc")
	globalFree                 = kernel32.NewProc("GlobalFree")
)

const (
	CF_DIB        = 8  // Device Independent Bitmap
	CF_DIBV5      = 17 // Device Independent Bitmap v5
	CF_BITMAP     = 2  // Bitmap handle
	GMEM_MOVEABLE = 0x0002
)

// BITMAPINFOHEADER structure for DIB format
type bitmapInfoHeader struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	ImageSize     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

// ReadClipboardImage reads an image from Windows clipboard
// This function is only available on Windows
func ReadClipboardImage() (image.Image, error) {
	return readClipboardImage()
}

// readClipboardImage reads an image from Windows clipboard (internal)
func readClipboardImage() (image.Image, error) {
	// Open clipboard
	ret, _, _ := openClipboard.Call(0)
	if ret == 0 {
		return nil, fmt.Errorf("failed to open clipboard")
	}
	defer closeClipboard.Call()

	// Check if DIB format is available
	ret, _, _ = isClipboardFormatAvailable.Call(CF_DIB)
	if ret == 0 {
		// Try DIBV5
		ret, _, _ = isClipboardFormatAvailable.Call(CF_DIBV5)
		if ret == 0 {
			return nil, fmt.Errorf("no image format available in clipboard")
		}
	}

	// Get clipboard data handle
	handle, _, err := getClipboardData.Call(CF_DIB)
	if handle == 0 {
		return nil, fmt.Errorf("failed to get clipboard data: %v", err)
	}

	// Lock the memory
	ptr, _, err := globalLock.Call(handle)
	if ptr == 0 {
		return nil, fmt.Errorf("failed to lock memory: %v", err)
	}
	defer globalUnlock.Call(handle)

	// Get size of clipboard data
	size, _, _ := globalSize.Call(handle)
	if size == 0 {
		return nil, fmt.Errorf("invalid clipboard data size")
	}

	// Read the data
	data := make([]byte, size)
	copy(data, (*[1 << 30]byte)(unsafe.Pointer(ptr))[:size:size])

	// Parse BITMAPINFOHEADER
	if len(data) < 40 {
		return nil, fmt.Errorf("clipboard data too short")
	}

	header := &bitmapInfoHeader{}
	reader := bytes.NewReader(data[:40])
	if err := binary.Read(reader, binary.LittleEndian, header); err != nil {
		return nil, fmt.Errorf("failed to read bitmap header: %v", err)
	}

	// Create image from DIB data
	img, err := dibToImage(data, header)
	if err != nil {
		return nil, fmt.Errorf("failed to convert DIB to image: %v", err)
	}

	return img, nil
}

// dibToImage converts DIB (Device Independent Bitmap) data to Go image.Image
func dibToImage(data []byte, header *bitmapInfoHeader) (image.Image, error) {
	width := int(header.Width)
	height := int(header.Height)
	if height < 0 {
		height = -height // Top-down DIB
	}

	// Calculate offset to pixel data
	offset := 40 + int(header.ClrUsed)*4 // Header size + color table

	// Create RGBA image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Calculate row size (must be aligned to 4 bytes)
	rowSize := ((width*int(header.BitCount) + 31) / 32) * 4

	if len(data) < offset+rowSize*height {
		return nil, fmt.Errorf("insufficient data for image")
	}

	// Copy pixel data (DIB is stored bottom-up, we need top-down)
	pixelData := data[offset:]

	if header.BitCount == 32 {
		// 32-bit RGBA
		for y := 0; y < height; y++ {
			srcY := height - 1 - y // Flip vertically
			for x := 0; x < width; x++ {
				idx := srcY*rowSize + x*4
				if idx+3 < len(pixelData) {
					// BGR to RGB conversion (Windows stores as BGR)
					img.Set(x, y, color.RGBA{
						R: pixelData[idx+2],
						G: pixelData[idx+1],
						B: pixelData[idx],
						A: pixelData[idx+3],
					})
				}
			}
		}
	} else if header.BitCount == 24 {
		// 24-bit RGB
		for y := 0; y < height; y++ {
			srcY := height - 1 - y // Flip vertically
			for x := 0; x < width; x++ {
				idx := srcY*rowSize + x*3
				if idx+2 < len(pixelData) {
					// BGR to RGB conversion
					img.Set(x, y, color.RGBA{
						R: pixelData[idx+2],
						G: pixelData[idx+1],
						B: pixelData[idx],
						A: 255,
					})
				}
			}
		}
	} else {
		return nil, fmt.Errorf("unsupported bit depth: %d", header.BitCount)
	}

	return img, nil
}

// imageToPNG converts image.Image to PNG bytes
func imageToPNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// WriteClipboardImage writes an image to Windows clipboard
// This function is only available on Windows
func WriteClipboardImage(img image.Image) error {
	return writeClipboardImage(img)
}

// writeClipboardImage writes an image to Windows clipboard (internal)
func writeClipboardImage(img image.Image) error {
	// Convert image to DIB format
	dibData, err := imageToDIB(img)
	if err != nil {
		return fmt.Errorf("failed to convert image to DIB: %v", err)
	}

	// Open clipboard
	ret, _, _ := openClipboard.Call(0)
	if ret == 0 {
		return fmt.Errorf("failed to open clipboard")
	}
	defer closeClipboard.Call()

	// Empty clipboard
	ret, _, _ = emptyClipboard.Call()
	if ret == 0 {
		return fmt.Errorf("failed to empty clipboard")
	}

	// Allocate global memory for DIB data
	handle, _, err := globalAlloc.Call(GMEM_MOVEABLE, uintptr(len(dibData)))
	if handle == 0 {
		return fmt.Errorf("failed to allocate global memory: %v", err)
	}

	// Lock memory
	ptr, _, err := globalLock.Call(handle)
	if ptr == 0 {
		// Free memory on error to prevent memory leak
		globalFree.Call(handle)
		return fmt.Errorf("failed to lock memory: %v", err)
	}

	// Copy data to global memory
	dst := (*[1 << 30]byte)(unsafe.Pointer(ptr))[:len(dibData):len(dibData)]
	copy(dst, dibData)

	// Unlock memory
	globalUnlock.Call(handle)

	// Set clipboard data
	ret, _, err = setClipboardData.Call(CF_DIB, handle)
	if ret == 0 {
		// Free memory on error to prevent memory leak
		globalFree.Call(handle)
		return fmt.Errorf("failed to set clipboard data: %v", err)
	}

	// Clipboard now owns the handle, don't free it
	return nil
}

// imageToDIB converts image.Image to DIB (Device Independent Bitmap) format
func imageToDIB(img image.Image) ([]byte, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Get image origin offset (some images don't start at 0,0)
	minX := bounds.Min.X
	minY := bounds.Min.Y

	// Validate dimensions
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid image dimensions: %dx%d", width, height)
	}

	// Create BITMAPINFOHEADER
	header := bitmapInfoHeader{
		Size:          40,            // Size of header
		Width:         int32(width),  // Image width
		Height:        int32(height), // Image height (positive = bottom-up)
		Planes:        1,             // Number of color planes
		BitCount:      32,            // Bits per pixel (32 = RGBA)
		Compression:   0,             // BI_RGB = 0 (no compression)
		ImageSize:     0,             // Image size (can be 0 for BI_RGB)
		XPelsPerMeter: 0,             // Horizontal resolution (optional)
		YPelsPerMeter: 0,             // Vertical resolution (optional)
		ClrUsed:       0,             // Number of colors used (0 = use BitCount)
		ClrImportant:  0,             // Important colors (0 = all)
	}

	// Calculate row size (must be aligned to 4 bytes)
	rowSize := ((width*4 + 3) / 4) * 4

	// Calculate image data size
	imageSize := rowSize * height
	header.ImageSize = uint32(imageSize)

	// Create buffer for DIB data
	var buf bytes.Buffer

	// Write header
	if err := binary.Write(&buf, binary.LittleEndian, header); err != nil {
		return nil, fmt.Errorf("failed to write header: %v", err)
	}

	// Write pixel data (bottom-up, BGR format)
	pixelData := make([]byte, imageSize)
	for y := 0; y < height; y++ {
		dstY := height - 1 - y // Flip vertically (DIB is bottom-up)
		for x := 0; x < width; x++ {
			// Get pixel color - use actual image coordinates (with offset)
			r, g, b, a := img.At(minX+x, minY+y).RGBA()

			// Convert from 16-bit to 8-bit
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)
			a8 := uint8(a >> 8)

			// DIB stores pixels as BGRA (not RGBA)
			idx := dstY*rowSize + x*4
			pixelData[idx] = b8   // Blue
			pixelData[idx+1] = g8 // Green
			pixelData[idx+2] = r8 // Red
			pixelData[idx+3] = a8 // Alpha
		}
	}

	// Write pixel data
	if _, err := buf.Write(pixelData); err != nil {
		return nil, fmt.Errorf("failed to write pixel data: %v", err)
	}

	return buf.Bytes(), nil
}
