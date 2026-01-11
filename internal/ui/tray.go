package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// SetupSystemTray creates a system tray icon with menu
func SetupSystemTray(app *App) {
	if desk, ok := app.fyneApp.(desktop.App); ok {
		appIcon := app.fyneApp.Icon()

		if appIcon != nil {
			desk.SetSystemTrayIcon(appIcon)
		}

		menu := fyne.NewMenu("",
			fyne.NewMenuItem("Aç", func() {
				app.Show()
			}),
			fyne.NewMenuItem("Gizle", func() {
				app.Hide()
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Çıkış", func() {
				app.fyneApp.Quit()
			}),
		)

		desk.SetSystemTrayMenu(menu)
	}
}
