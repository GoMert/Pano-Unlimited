package ui

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"pano/internal/clipboard"
	"pano/internal/storage"
)

// ClipboardList displays clipboard items
type ClipboardList struct {
	widget.BaseWidget
	manager  *clipboard.Manager
	items    []storage.ClipboardItem
	onSelect func(id string)
	onPin    func(id string)
	onDelete func(id string)
}

// NewClipboardList creates a new clipboard list widget
func NewClipboardList(manager *clipboard.Manager) *ClipboardList {
	list := &ClipboardList{
		manager: manager,
		items:   []storage.ClipboardItem{},
	}
	list.ExtendBaseWidget(list)
	return list
}

// SetCallbacks sets the callback functions
func (c *ClipboardList) SetCallbacks(onSelect, onPin, onDelete func(id string)) {
	c.onSelect = onSelect
	c.onPin = onPin
	c.onDelete = onDelete
}

// Refresh updates the list with current items
func (c *ClipboardList) Refresh() {
	c.items = c.manager.GetAllItems()
	c.BaseWidget.Refresh()
}

// CreateRenderer creates the widget renderer
func (c *ClipboardList) CreateRenderer() fyne.WidgetRenderer {
	return &clipboardListRenderer{
		list: c,
	}
}

type clipboardListRenderer struct {
	list      *ClipboardList
	container *fyne.Container
}

func (r *clipboardListRenderer) Layout(size fyne.Size) {
	if r.container != nil {
		r.container.Resize(size)
	}
}

func (r *clipboardListRenderer) MinSize() fyne.Size {
	if r.container != nil {
		return r.container.MinSize()
	}
	return fyne.NewSize(400, 300)
}

func (r *clipboardListRenderer) Refresh() {
	r.container = r.buildList()
	r.container.Refresh()
}

func (r *clipboardListRenderer) Objects() []fyne.CanvasObject {
	if r.container == nil {
		r.container = r.buildList()
	}
	return []fyne.CanvasObject{r.container}
}

func (r *clipboardListRenderer) Destroy() {}

func (r *clipboardListRenderer) buildList() *fyne.Container {
	items := []fyne.CanvasObject{}

	if len(r.list.items) == 0 {
		// Empty state
		emptyLabel := widget.NewLabel("Pano geçmişi boş")
		emptyLabel.Alignment = fyne.TextAlignCenter
		
		hintLabel := widget.NewLabel("Bir şey kopyaladığınızda burada görünecek")
		hintLabel.Alignment = fyne.TextAlignCenter
		
		items = append(items, container.NewVBox(
			widget.NewSeparator(),
			emptyLabel,
			hintLabel,
			widget.NewSeparator(),
		))
	} else {
		for _, item := range r.list.items {
			items = append(items, r.createItemCard(item))
		}
	}

	return container.NewVBox(items...)
}

func (r *clipboardListRenderer) createItemCard(item storage.ClipboardItem) fyne.CanvasObject {
	var previewContent fyne.CanvasObject
	var typeLabel string

	if item.Type == "text" {
		typeLabel = "METİN"
		// Get decrypted content for preview
		content, err := r.list.manager.GetItemContent(item.ID)
		preview := ""
		if err == nil {
			preview = string(content)
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			// Clean up for preview
			preview = strings.ReplaceAll(preview, "\r\n", " ")
			preview = strings.ReplaceAll(preview, "\n", " ")
			preview = strings.ReplaceAll(preview, "\r", " ")
			preview = strings.TrimSpace(preview)
		}
		previewLabel := widget.NewLabel(preview)
		previewLabel.Wrapping = fyne.TextWrapWord
		previewContent = previewLabel
	} else if item.Type == "image" {
		typeLabel = "GÖRSEL"
		// Get image content and create thumbnail
		content, err := r.list.manager.GetItemContent(item.ID)
		if err == nil {
			img, err := png.Decode(bytes.NewReader(content))
			if err == nil {
				// Larger thumbnail for better quality (300x200)
				thumbnail := createThumbnail(img, 300, 200)
				imgCanvas := canvas.NewImageFromImage(thumbnail)
				imgCanvas.FillMode = canvas.ImageFillContain
				imgCanvas.SetMinSize(fyne.NewSize(280, 180))
				previewContent = imgCanvas
			} else {
				previewContent = widget.NewLabel("Görsel yüklenemedi")
			}
		} else {
			previewContent = widget.NewLabel("Görsel yüklenemedi")
		}
	} else {
		typeLabel = "BİLİNMEYEN"
		previewContent = widget.NewLabel("Desteklenmeyen içerik türü")
	}

	// Create info line
	timestamp := formatTimestamp(item.Timestamp)
	sizeStr := formatSize(item.Size)
	
	// Type badge
	typeBadge := widget.NewLabel(typeLabel)
	typeBadge.TextStyle = fyne.TextStyle{Bold: true}

	// Pinned indicator
	var pinnedBadge *widget.Label
	if item.Pinned {
		pinnedBadge = widget.NewLabel("[SABİT]")
		pinnedBadge.TextStyle = fyne.TextStyle{Bold: true}
	}

	// Info text
	infoText := widget.NewLabel(fmt.Sprintf("%s  •  %s", sizeStr, timestamp))

	// Header row
	var headerRow *fyne.Container
	if item.Pinned {
		headerRow = container.NewHBox(typeBadge, pinnedBadge, widget.NewLabel(""), infoText)
	} else {
		headerRow = container.NewBorder(nil, nil, typeBadge, infoText)
	}

	// Action buttons - clean text
	copyBtn := widget.NewButton("Kopyala", func() {
		if r.list.onSelect != nil {
			r.list.onSelect(item.ID)
		}
	})
	copyBtn.Importance = widget.HighImportance

	pinBtnText := "Sabitle"
	if item.Pinned {
		pinBtnText = "Kaldır"
	}
	pinBtn := widget.NewButton(pinBtnText, func() {
		if r.list.onPin != nil {
			r.list.onPin(item.ID)
		}
	})

	deleteBtn := widget.NewButton("Sil", func() {
		if r.list.onDelete != nil {
			r.list.onDelete(item.ID)
		}
	})
	deleteBtn.Importance = widget.DangerImportance

	// Button row
	buttonRow := container.NewHBox(copyBtn, pinBtn, deleteBtn)

	// Card content
	cardContent := container.NewVBox(
		headerRow,
		previewContent,
		buttonRow,
	)

	// Background colors - theme aware
	bgColor := GetCardBackgroundColor(item.Pinned)
	borderColor := GetCardBorderColor()
	
	bg := canvas.NewRectangle(bgColor)
	bg.CornerRadius = 6
	bg.StrokeWidth = 1
	bg.StrokeColor = borderColor

	// Card with padding
	paddedCard := container.NewPadded(cardContent)

	return container.NewVBox(
		container.NewStack(bg, paddedCard),
	)
}

// createThumbnail creates a high-quality thumbnail using bilinear interpolation
func createThumbnail(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	scaleX := float64(maxWidth) / float64(width)
	scaleY := float64(maxHeight) / float64(height)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	if scale >= 1.0 {
		return img // Return original if smaller than max
	}

	newWidth := int(float64(width) * scale)
	newHeight := int(float64(height) * scale)
	
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	thumbnail := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Bilinear interpolation for better quality
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			// Map to source coordinates
			srcX := float64(x) / scale
			srcY := float64(y) / scale

			// Get the four nearest pixels
			x0 := int(srcX)
			y0 := int(srcY)
			x1 := x0 + 1
			y1 := y0 + 1

			// Clamp to image bounds
			if x0 < 0 {
				x0 = 0
			}
			if y0 < 0 {
				y0 = 0
			}
			if x1 >= width {
				x1 = width - 1
			}
			if y1 >= height {
				y1 = height - 1
			}

			// Calculate interpolation weights
			xWeight := srcX - float64(x0)
			yWeight := srcY - float64(y0)

			// Get colors of four surrounding pixels
			c00 := img.At(bounds.Min.X+x0, bounds.Min.Y+y0)
			c10 := img.At(bounds.Min.X+x1, bounds.Min.Y+y0)
			c01 := img.At(bounds.Min.X+x0, bounds.Min.Y+y1)
			c11 := img.At(bounds.Min.X+x1, bounds.Min.Y+y1)

			// Convert to RGBA
			r00, g00, b00, a00 := c00.RGBA()
			r10, g10, b10, a10 := c10.RGBA()
			r01, g01, b01, a01 := c01.RGBA()
			r11, g11, b11, a11 := c11.RGBA()

			// Bilinear interpolation
			r := bilinearInterp(r00, r10, r01, r11, xWeight, yWeight)
			g := bilinearInterp(g00, g10, g01, g11, xWeight, yWeight)
			b := bilinearInterp(b00, b10, b01, b11, xWeight, yWeight)
			a := bilinearInterp(a00, a10, a01, a11, xWeight, yWeight)

			thumbnail.Set(x, y, color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			})
		}
	}

	return thumbnail
}

// bilinearInterp performs bilinear interpolation
func bilinearInterp(c00, c10, c01, c11 uint32, xWeight, yWeight float64) uint32 {
	// Interpolate along x for top and bottom
	top := float64(c00)*(1-xWeight) + float64(c10)*xWeight
	bottom := float64(c01)*(1-xWeight) + float64(c11)*xWeight
	// Interpolate along y
	return uint32(top*(1-yWeight) + bottom*yWeight)
}

func formatSize(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatTimestamp(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "Az önce"
	} else if diff < time.Hour {
		return fmt.Sprintf("%d dk önce", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%d saat önce", int(diff.Hours()))
	} else if diff < 7*24*time.Hour {
		return fmt.Sprintf("%d gün önce", int(diff.Hours()/24))
	}
	return t.Format("02.01.2006")
}
