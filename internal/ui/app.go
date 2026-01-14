package ui

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"pano/internal/clipboard"
	"pano/internal/storage"
	"pano/internal/system"
)

type App struct {
	fyneApp     fyne.App
	window      fyne.Window
	manager     *clipboard.Manager
	monitor     *clipboard.Monitor
	list        *ClipboardList
	autostart   *system.AutostartManager
	isVisible   bool
	statusLabel *widget.Label
	isDarkMode  bool
	toastMu     sync.Mutex
}

func NewApp(fyneApp fyne.App, db *storage.Database, autostart *system.AutostartManager) *App {
	app := &App{
		fyneApp:   fyneApp,
		manager:   clipboard.NewManager(db),
		monitor:   clipboard.NewMonitor(db),
		autostart: autostart,
		isVisible: false,
	}

	app.isDarkMode = fyneApp.Preferences().BoolWithFallback("dark_mode", true)

	// Load saved max items limit
	savedLimit := fyneApp.Preferences().IntWithFallback("max_items", 100)
	app.manager.SetMaxItems(savedLimit)

	if app.isDarkMode {
		fyneApp.Settings().SetTheme(NewDarkTheme())
	} else {
		fyneApp.Settings().SetTheme(NewLightTheme())
	}

	app.window = fyneApp.NewWindow("Pano")
	app.window.Resize(fyne.NewSize(380, 520))
	app.window.CenterOnScreen()

	app.buildUI()

	app.window.SetCloseIntercept(func() {
		app.Hide()
	})

	// Set limit warning callback on monitor
	app.monitor.SetOnLimitWarn(func(remaining int) {
		if remaining == 0 {
			app.sendNotification("Limit Doldu", "Pano limiti doldu! Yeni kopyalamalar kaydedilmiyor.")
		} else {
			app.sendNotification("Pano Uyarısı", fmt.Sprintf("Sadece %d alan kaldı! Yakında kopyaladıkların kaydedilmeyecek.", remaining))
		}
	})

	app.monitor.SetOnChange(func(itemType string, content []byte) {
		app.list.Refresh()
		app.updateStatus()
	})

	return app
}

func (a *App) sendNotification(title, message string) {
	notification := fyne.NewNotification(title, message)
	a.fyneApp.SendNotification(notification)
}

func (a *App) buildUI() {
	a.list = NewClipboardList(a.manager)

	a.list.SetCallbacks(
		func(id string) {
			if err := a.manager.CopyToClipboard(id); err != nil {
				dialog.ShowError(err, a.window)
			} else {
				a.showToast("Panoya kopyalandı")
			}
		},
		func(id string) {
			if err := a.manager.PinItem(id); err != nil {
				dialog.ShowError(err, a.window)
			} else {
				a.list.Refresh()
				a.updateStatus()
			}
		},
		func(id string) {
			if err := a.manager.DeleteItem(id); err != nil {
				dialog.ShowError(err, a.window)
			} else {
				a.list.Refresh()
				a.updateStatus()
			}
		},
	)

	titleLabel := widget.NewLabelWithStyle("Pano Geçmişi", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		a.list.Refresh()
		a.updateStatus()
		a.showToast("Yenilendi")
	})

	settingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		a.showSettingsDialog()
	})

	clearBtn := widget.NewButtonWithIcon("Temizle", theme.DeleteIcon(), func() {
		a.showClearAllDialog()
	})
	clearBtn.Importance = widget.DangerImportance

	header := container.NewBorder(nil, nil, titleLabel, container.NewHBox(refreshBtn, settingsBtn, clearBtn))

	a.statusLabel = widget.NewLabel("")
	a.updateStatus()

	shortcutLabel := widget.NewLabelWithStyle("Ctrl+Shift+V", fyne.TextAlignTrailing, fyne.TextStyle{Italic: true})

	footer := container.NewBorder(nil, nil, a.statusLabel, shortcutLabel)

	scroll := container.NewVScroll(a.list)

	content := container.NewBorder(
		container.NewVBox(header, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), footer),
		nil, nil,
		scroll,
	)

	a.window.SetContent(container.NewPadded(content))
}

func (a *App) showToast(message string) {
	a.toastMu.Lock()
	defer a.toastMu.Unlock()
	
	a.statusLabel.SetText("[OK] " + message)
	go func() {
		time.Sleep(1500 * time.Millisecond)
		a.toastMu.Lock()
		a.updateStatusInternal()
		a.toastMu.Unlock()
	}()
}

func (a *App) updateStatus() {
	a.toastMu.Lock()
	defer a.toastMu.Unlock()
	a.updateStatusInternal()
}

func (a *App) updateStatusInternal() {
	total := a.manager.GetItemCount()
	maxItems := a.manager.GetMaxItems()
	pinned := a.manager.GetPinnedCount()
	a.statusLabel.SetText(fmt.Sprintf("%d/%d öğe - %d sabit", total, maxItems, pinned))
}

func (a *App) showSettingsDialog() {
	isEnabled, err := a.autostart.IsEnabled()
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	// Theme selection
	themeLabel := widget.NewLabelWithStyle("Tema", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	themeSelect := widget.NewSelect([]string{"Koyu Tema", "Açık Tema"}, func(s string) {
		if s == "Koyu Tema" {
			a.isDarkMode = true
			a.fyneApp.Settings().SetTheme(NewDarkTheme())
		} else {
			a.isDarkMode = false
			a.fyneApp.Settings().SetTheme(NewLightTheme())
		}
		a.fyneApp.Preferences().SetBool("dark_mode", a.isDarkMode)
		a.list.Refresh()
	})
	if a.isDarkMode {
		themeSelect.SetSelected("Koyu Tema")
	} else {
		themeSelect.SetSelected("Açık Tema")
	}

	// Max items limit
	limitLabel := widget.NewLabelWithStyle("Maksimum Öğe Sayısı", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	currentLimit := a.manager.GetMaxItems()
	limitValue := widget.NewLabel(fmt.Sprintf("%d öğe", currentLimit))
	
	limitSlider := widget.NewSlider(10, 500)
	limitSlider.Step = 10
	limitSlider.Value = float64(currentLimit)
	limitSlider.OnChanged = func(v float64) {
		limitValue.SetText(fmt.Sprintf("%d öğe", int(v)))
	}
	limitSlider.OnChangeEnded = func(v float64) {
		newLimit := int(v)
		a.manager.SetMaxItems(newLimit)
		a.fyneApp.Preferences().SetInt("max_items", newLimit)
		a.updateStatus()
	}

	// Autostart
	autostartLabel := widget.NewLabelWithStyle("Başlangıç", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	autostartCheck := widget.NewCheck("Windows ile başlat", func(checked bool) {
		if checked {
			if err := a.autostart.Enable(); err != nil {
				dialog.ShowError(err, a.window)
			}
		} else {
			if err := a.autostart.Disable(); err != nil {
				dialog.ShowError(err, a.window)
			}
		}
	})
	autostartCheck.Checked = isEnabled

	// Info
	infoLabel := widget.NewLabelWithStyle("Hakkında", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	infoText := widget.NewLabel("Kısayol: Ctrl+Shift+V\nŞifreleme: AES-256")

	dialogContent := container.NewVBox(
		themeLabel,
		themeSelect,
		widget.NewSeparator(),
		limitLabel,
		container.NewBorder(nil, nil, nil, limitValue, limitSlider),
		widget.NewSeparator(),
		autostartLabel,
		autostartCheck,
		widget.NewSeparator(),
		infoLabel,
		infoText,
	)

	dialog.ShowCustom("Ayarlar", "Kapat", dialogContent, a.window)
}

func (a *App) showClearAllDialog() {
	count := a.manager.GetItemCount()
	if count == 0 {
		dialog.ShowInformation("Bilgi", "Silinecek öğe yok.", a.window)
		return
	}

	dialog.ShowConfirm("Tümünü Temizle",
		fmt.Sprintf("%d öğe silinecek. Devam edilsin mi?", count),
		func(ok bool) {
			if ok {
				if err := a.manager.ClearAll(); err != nil {
					dialog.ShowError(err, a.window)
				} else {
					thumbCache.clear()
					a.list.Refresh()
					a.updateStatus()
				}
			}
		}, a.window)
}

func (a *App) Show() {
	a.isVisible = true
	a.list.Refresh()
	a.updateStatus()
	a.window.Show()
	a.window.RequestFocus()
	BringWindowToFront("Pano")
}

func (a *App) Hide() {
	a.isVisible = false
	a.window.Hide()
}

func (a *App) Toggle() {
	if a.isVisible {
		a.Hide()
	} else {
		a.Show()
	}
}

func (a *App) StartMonitoring() error {
	return a.monitor.Start()
}

func (a *App) StopMonitoring() {
	a.monitor.Stop()
}

func (a *App) Run() {
	a.isVisible = true
	a.window.ShowAndRun()
}

func (a *App) GetWindow() fyne.Window {
	return a.window
}

func (a *App) GetFyneApp() fyne.App {
	return a.fyneApp
}
