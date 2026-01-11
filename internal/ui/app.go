package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"pano/internal/clipboard"
	"pano/internal/storage"
	"pano/internal/system"
)

// App represents the main application window
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
}

// NewApp creates a new application
func NewApp(fyneApp fyne.App, db *storage.Database, autostart *system.AutostartManager) *App {
	app := &App{
		fyneApp:   fyneApp,
		manager:   clipboard.NewManager(db),
		monitor:   clipboard.NewMonitor(db),
		autostart: autostart,
	}

	// Load theme preference
	app.isDarkMode = fyneApp.Preferences().BoolWithFallback("dark_mode", false)
	
	// Set theme based on preference
	if app.isDarkMode {
		fyneApp.Settings().SetTheme(NewDarkTheme())
	} else {
		fyneApp.Settings().SetTheme(NewLightTheme())
	}

	// Create window
	app.window = fyneApp.NewWindow("Pano")
	app.window.Resize(fyne.NewSize(520, 700))
	app.window.CenterOnScreen()

	// Create UI
	app.buildUI()

	// Hide instead of quit on close
	app.window.SetCloseIntercept(func() {
		app.Hide()
	})

	app.isVisible = false

	// Set up clipboard monitor callback
	app.monitor.SetOnChange(func(itemType string, content []byte) {
		app.list.Refresh()
	})

	return app
}

// buildUI constructs the user interface
func (a *App) buildUI() {
	// Create clipboard list
	a.list = NewClipboardList(a.manager)

	// Set callbacks
	a.list.SetCallbacks(
		func(id string) {
			defer func() {
				if r := recover(); r != nil {
					dialog.ShowError(fmt.Errorf("Kopyalama hatası: %v", r), a.window)
				}
			}()
			if err := a.manager.CopyToClipboard(id); err != nil {
				dialog.ShowError(err, a.window)
			} else {
				a.updateStatus()
			}
		},
		func(id string) {
			defer func() {
				if r := recover(); r != nil {
					dialog.ShowError(fmt.Errorf("Sabitleme hatası: %v", r), a.window)
				}
			}()
			if err := a.manager.PinItem(id); err != nil {
				dialog.ShowError(err, a.window)
			} else {
				a.list.Refresh()
				a.updateStatus()
			}
		},
		func(id string) {
			defer func() {
				if r := recover(); r != nil {
					dialog.ShowError(fmt.Errorf("Silme hatası: %v", r), a.window)
				}
			}()
			if err := a.manager.DeleteItem(id); err != nil {
				dialog.ShowError(err, a.window)
			} else {
				a.list.Refresh()
				a.updateStatus()
			}
		},
	)

	// Header
	titleLabel := widget.NewLabel("Pano")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Toolbar buttons
	settingsBtn := widget.NewButton("Ayarlar", func() {
		a.showSettingsDialog()
	})

	refreshBtn := widget.NewButton("Yenile", func() {
		a.list.Refresh()
		a.updateStatus()
	})

	clearAllBtn := widget.NewButton("Tümünü Sil", func() {
		a.showClearAllDialog()
	})
	clearAllBtn.Importance = widget.DangerImportance

	// Toolbar layout
	toolbar := container.NewHBox(
		settingsBtn,
		refreshBtn,
		layout.NewSpacer(),
		clearAllBtn,
	)

	// Status bar
	a.statusLabel = widget.NewLabel("")
	a.updateStatus()

	shortcutLabel := widget.NewLabel("Ctrl+Shift+V")

	// Header section
	header := container.NewVBox(
		container.NewBorder(nil, nil, titleLabel, nil),
		toolbar,
		widget.NewSeparator(),
	)

	// Footer section
	footer := container.NewVBox(
		widget.NewSeparator(),
		container.NewBorder(nil, nil, a.statusLabel, shortcutLabel),
	)

	// Main layout
	content := container.NewBorder(
		header,
		footer,
		nil,
		nil,
		container.NewScroll(a.list),
	)

	a.window.SetContent(content)
}

// updateStatus updates the status bar
func (a *App) updateStatus() {
	total := a.manager.GetItemCount()
	pinned := a.manager.GetPinnedCount()
	a.statusLabel.SetText(fmt.Sprintf("%d öğe  •  %d sabit", total, pinned))
}

// showSettingsDialog shows settings dialog
func (a *App) showSettingsDialog() {
	isEnabled, err := a.autostart.IsEnabled()
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	// Status text
	statusText := "Kapalı"
	if isEnabled {
		statusText = "Açık"
	}

	// Title
	titleLabel := widget.NewLabel("Ayarlar")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Theme section
	themeTitle := widget.NewLabel("Tema")
	themeTitle.TextStyle = fyne.TextStyle{Bold: true}

	themeStatus := widget.NewLabel("Durum: Açık Tema")
	if a.isDarkMode {
		themeStatus.SetText("Durum: Koyu Tema")
	}

	var themeBtn *widget.Button
	themeBtn = widget.NewButton("", nil)

	updateThemeUI := func() {
		if a.isDarkMode {
			themeStatus.SetText("Durum: Koyu Tema")
			themeBtn.SetText("Açık Tema")
		} else {
			themeStatus.SetText("Durum: Açık Tema")
			themeBtn.SetText("Koyu Tema")
		}
		themeBtn.Refresh()
	}

	if a.isDarkMode {
		themeBtn.SetText("Açık Tema")
	} else {
		themeBtn.SetText("Koyu Tema")
	}

	themeBtn.OnTapped = func() {
		a.isDarkMode = !a.isDarkMode
		a.fyneApp.Preferences().SetBool("dark_mode", a.isDarkMode)
		
		if a.isDarkMode {
			a.fyneApp.Settings().SetTheme(NewDarkTheme())
		} else {
			a.fyneApp.Settings().SetTheme(NewLightTheme())
		}
		
		updateThemeUI()
		a.list.Refresh()
	}

	// Autostart section
	autostartTitle := widget.NewLabel("Başlangıç Ayarları")
	autostartTitle.TextStyle = fyne.TextStyle{Bold: true}

	autostartStatus := widget.NewLabel(fmt.Sprintf("Durum: %s", statusText))

	var autostartBtn *widget.Button
	autostartBtn = widget.NewButton("", nil)

	updateAutostartUI := func() {
		enabled, _ := a.autostart.IsEnabled()
		if enabled {
			autostartStatus.SetText("Durum: Açık")
			autostartBtn.SetText("Kapat")
			autostartBtn.Importance = widget.WarningImportance
		} else {
			autostartStatus.SetText("Durum: Kapalı")
			autostartBtn.SetText("Aç")
			autostartBtn.Importance = widget.HighImportance
		}
		autostartBtn.Refresh()
	}

	if isEnabled {
		autostartBtn.SetText("Kapat")
		autostartBtn.Importance = widget.WarningImportance
	} else {
		autostartBtn.SetText("Aç")
		autostartBtn.Importance = widget.HighImportance
	}

	autostartBtn.OnTapped = func() {
		currentEnabled, _ := a.autostart.IsEnabled()
		if currentEnabled {
			dialog.ShowConfirm(
				"Uyarı",
				"Başlangıçta çalıştırmayı kapatmak istediğinizden emin misiniz?",
				func(confirm bool) {
					if confirm {
						if err := a.autostart.Disable(); err != nil {
							dialog.ShowError(err, a.window)
						} else {
							updateAutostartUI()
							dialog.ShowInformation("Başarılı", "Başlangıçta çalıştırma kapatıldı.", a.window)
						}
					}
				},
				a.window,
			)
		} else {
			if err := a.autostart.Enable(); err != nil {
				dialog.ShowError(err, a.window)
			} else {
				updateAutostartUI()
				dialog.ShowInformation("Başarılı", "Başlangıçta çalıştırma açıldı.", a.window)
			}
		}
	}

	// Info section
	infoTitle := widget.NewLabel("Bilgi")
	infoTitle.TextStyle = fyne.TextStyle{Bold: true}

	infoText := widget.NewLabel(
		"Kısayol: Ctrl+Shift+V\n" +
			"Maksimum: 100 öğe\n" +
			"Veriler şifrelenmiş olarak saklanır")
	infoText.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		themeTitle,
		themeStatus,
		themeBtn,
		widget.NewSeparator(),
		autostartTitle,
		autostartStatus,
		autostartBtn,
		widget.NewSeparator(),
		infoTitle,
		infoText,
	)

	dialog.ShowCustom("Ayarlar", "Kapat", content, a.window)
}

// showClearAllDialog shows confirmation dialog for clearing all items
func (a *App) showClearAllDialog() {
	itemCount := a.manager.GetItemCount()
	pinnedCount := a.manager.GetPinnedCount()

	if itemCount == 0 {
		dialog.ShowInformation("Bilgi", "Silinecek öğe yok.", a.window)
		return
	}

	dialog.ShowConfirm(
		"Tümünü Sil",
		fmt.Sprintf("%d öğe silinecek.\n(%d tanesi sabitlenmiş)\n\nDevam edilsin mi?", itemCount, pinnedCount),
		func(firstConfirm bool) {
			if firstConfirm {
				dialog.ShowConfirm(
					"Son Onay",
					"Bu işlem geri alınamaz.\n\nEmin misiniz?",
					func(secondConfirm bool) {
						if secondConfirm {
							if err := a.manager.ClearAll(); err != nil {
								dialog.ShowError(err, a.window)
							} else {
								a.list.Refresh()
								a.updateStatus()
								dialog.ShowInformation("Başarılı", "Tüm öğeler silindi.", a.window)
							}
						}
					},
					a.window,
				)
			}
		},
		a.window,
	)
}

// Show displays the window
func (a *App) Show() {
	a.list.Refresh()
	a.updateStatus()
	a.window.Show()
	a.window.RequestFocus()
	BringWindowToFront("Pano")
	a.isVisible = true
}

// Hide hides the window
func (a *App) Hide() {
	a.window.Hide()
	a.isVisible = false
}

// Toggle toggles window visibility
func (a *App) Toggle() {
	if a.isVisible {
		a.Hide()
	} else {
		a.Show()
	}
}

// StartMonitoring starts clipboard monitoring
func (a *App) StartMonitoring() error {
	return a.monitor.Start()
}

// Run runs the application
func (a *App) Run() {
	a.isVisible = false
	a.window.ShowAndRun()
}

// GetWindow returns the application window
func (a *App) GetWindow() fyne.Window {
	return a.window
}

// GetFyneApp returns the Fyne application instance
func (a *App) GetFyneApp() fyne.App {
	return a.fyneApp
}
