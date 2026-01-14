package ui

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"pano/internal/clipboard"
	"pano/internal/storage"
)

type thumbnailCache struct {
	mu    sync.RWMutex
	cache map[string]image.Image
}

var thumbCache = &thumbnailCache{
	cache: make(map[string]image.Image),
}

func (tc *thumbnailCache) get(id string) (image.Image, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	img, ok := tc.cache[id]
	return img, ok
}

func (tc *thumbnailCache) set(id string, img image.Image) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.cache[id] = img
}

func (tc *thumbnailCache) clear() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.cache = make(map[string]image.Image)
}

type ClipboardList struct {
	widget.BaseWidget
	manager  *clipboard.Manager
	items    []storage.ClipboardItem
	onSelect func(id string)
	onPin    func(id string)
	onDelete func(id string)
}

func NewClipboardList(manager *clipboard.Manager) *ClipboardList {
	list := &ClipboardList{
		manager: manager,
		items:   []storage.ClipboardItem{},
	}
	list.ExtendBaseWidget(list)
	return list
}

func (c *ClipboardList) SetCallbacks(onSelect, onPin, onDelete func(id string)) {
	c.onSelect = onSelect
	c.onPin = onPin
	c.onDelete = onDelete
}

func (c *ClipboardList) Refresh() {
	c.items = c.manager.GetAllItems()
	c.BaseWidget.Refresh()
}

func (c *ClipboardList) CreateRenderer() fyne.WidgetRenderer {
	return &clipboardListRenderer{list: c}
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
	return fyne.NewSize(300, 200)
}

func (r *clipboardListRenderer) Refresh() {
	r.container = r.buildList()
}

func (r *clipboardListRenderer) Objects() []fyne.CanvasObject {
	if r.container == nil {
		r.container = r.buildList()
	}
	return []fyne.CanvasObject{r.container}
}

func (r *clipboardListRenderer) Destroy() {}

func (r *clipboardListRenderer) buildList() *fyne.Container {
	if len(r.list.items) == 0 {
		return r.createEmptyState()
	}

	items := make([]fyne.CanvasObject, 0, len(r.list.items))
	for _, item := range r.list.items {
		items = append(items, r.createCard(item))
	}

	return container.NewVBox(items...)
}

func (r *clipboardListRenderer) createEmptyState() *fyne.Container {
	icon := widget.NewIcon(theme.ContentPasteIcon())
	title := widget.NewLabelWithStyle("Pano boş", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	hint := widget.NewLabelWithStyle("Bir şey kopyaladığınızda burada görünür", fyne.TextAlignCenter, fyne.TextStyle{})

	return container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(icon),
		container.NewCenter(title),
		container.NewCenter(hint),
		layout.NewSpacer(),
	)
}

func (r *clipboardListRenderer) createCard(item storage.ClipboardItem) fyne.CanvasObject {
	var content fyne.CanvasObject

	if item.Type == "text" {
		data, err := r.list.manager.GetItemContent(item.ID)
		text := ""
		if err == nil {
			text = string(data)
			text = strings.ReplaceAll(text, "\r\n", " ")
			text = strings.ReplaceAll(text, "\n", " ")
			text = strings.ReplaceAll(text, "\r", " ")
			text = strings.TrimSpace(text)
			if len(text) > 100 {
				text = text[:100] + "..."
			}
		}

		label := widget.NewLabel(text)
		label.Wrapping = fyne.TextWrapWord
		content = label

	} else if item.Type == "image" {
		var img image.Image
		if cached, ok := thumbCache.get(item.ID); ok {
			img = cached
		} else {
			data, err := r.list.manager.GetItemContent(item.ID)
			if err == nil {
				decoded, err := png.Decode(bytes.NewReader(data))
				if err == nil {
					img = createThumbnailFast(decoded, 320, 160)
					thumbCache.set(item.ID, img)
				}
			}
		}

		if img != nil {
			imgWidget := canvas.NewImageFromImage(img)
			imgWidget.FillMode = canvas.ImageFillContain
			imgWidget.ScaleMode = canvas.ImageScaleSmooth
			imgWidget.SetMinSize(fyne.NewSize(320, 140))
			content = container.NewCenter(imgWidget)
		} else {
			content = widget.NewLabel("Görsel yüklenemedi")
		}
	} else {
		content = widget.NewLabel("Bilinmeyen tür")
	}

	timeStr := formatTimestamp(item.Timestamp)
	sizeStr := formatSize(item.Size)

	var infoStr string
	if item.Type == "text" {
		infoStr = fmt.Sprintf("Metin - %s - %s", sizeStr, timeStr)
	} else {
		infoStr = fmt.Sprintf("Görsel - %s - %s", sizeStr, timeStr)
	}

	if item.Pinned {
		infoStr = "[Sabit] " + infoStr
	}

	infoLabel := widget.NewLabelWithStyle(infoStr, fyne.TextAlignLeading, fyne.TextStyle{Italic: true})

	itemID := item.ID
	
	copyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		if r.list.onSelect != nil {
			r.list.onSelect(itemID)
		}
	})
	copyBtn.Importance = widget.HighImportance

	pinBtn := widget.NewButtonWithIcon("", theme.CheckButtonIcon(), func() {
		if r.list.onPin != nil {
			r.list.onPin(itemID)
		}
	})
	if item.Pinned {
		pinBtn.Icon = theme.CheckButtonCheckedIcon()
	}

	delBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		if r.list.onDelete != nil {
			r.list.onDelete(itemID)
		}
	})

	buttons := container.NewHBox(copyBtn, pinBtn, delBtn)

	cardContent := container.NewVBox(
		content,
		container.NewBorder(nil, nil, infoLabel, buttons),
	)

	bg := canvas.NewRectangle(GetCardBackgroundColor(item.Pinned))
	bg.CornerRadius = 8
	bg.StrokeWidth = 1
	bg.StrokeColor = GetCardBorderColor(item.Pinned)

	card := container.NewStack(bg, container.NewPadded(cardContent))

	return card
}

// Fast thumbnail using nearest neighbor (much faster than bilinear)
func createThumbnailFast(img image.Image, maxW, maxH int) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	scaleX := float64(maxW) / float64(w)
	scaleY := float64(maxH) / float64(h)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	if scale >= 1.0 {
		return img
	}

	newW := int(float64(w) * scale)
	newH := int(float64(h) * scale)
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	thumb := image.NewRGBA(image.Rect(0, 0, newW, newH))

	for y := 0; y < newH; y++ {
		srcY := int(float64(y) / scale)
		if srcY >= h {
			srcY = h - 1
		}
		for x := 0; x < newW; x++ {
			srcX := int(float64(x) / scale)
			if srcX >= w {
				srcX = w - 1
			}
			c := img.At(bounds.Min.X+srcX, bounds.Min.Y+srcY)
			r, g, b, a := c.RGBA()
			thumb.Set(x, y, color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			})
		}
	}

	return thumb
}

func formatSize(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	kb := float64(bytes) / 1024
	if kb < 1024 {
		return fmt.Sprintf("%.1f KB", kb)
	}
	mb := kb / 1024
	return fmt.Sprintf("%.1f MB", mb)
}

func formatTimestamp(t time.Time) string {
	diff := time.Since(t)

	if diff < time.Minute {
		return "Az önce"
	}
	if diff < time.Hour {
		return fmt.Sprintf("%d dk", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%d sa", int(diff.Hours()))
	}
	if diff < 7*24*time.Hour {
		return fmt.Sprintf("%d gün", int(diff.Hours()/24))
	}
	return t.Format("02.01.2006")
}
