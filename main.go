package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"

	"pano/internal/storage"
	"pano/internal/system"
	"pano/internal/ui"
)

func main() {
	// Initialize Fyne app with ID
	fyneApp := app.NewWithID("com.pano.clipboard")

	// Set application icon for system tray
	appIcon := getPanoIcon()
	fyneApp.SetIcon(appIcon)

	// Initialize database
	db, err := storage.NewDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize autostart manager
	autostart, err := system.NewAutostartManager()
	if err != nil {
		log.Fatalf("Failed to initialize autostart: %v", err)
	}

	// Create UI
	appUI := ui.NewApp(fyneApp, db, autostart)

	// Setup system tray
	ui.SetupSystemTray(appUI)

	// Initialize hotkey manager (Ctrl+Shift+V to toggle window)
	hotkeyMgr := system.NewHotkeyManager()
	hotkeyMgr.SetCallback(func() {
		appUI.Toggle()
	})

	// Start hotkey listener
	if err := hotkeyMgr.Start(); err != nil {
		log.Printf("Warning: Failed to register hotkey: %v", err)
	}

	// Start clipboard monitoring
	if err := appUI.StartMonitoring(); err != nil {
		dialog.ShowError(err, appUI.GetWindow())
		return
	}

	// Setup graceful shutdown handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		hotkeyMgr.Stop()
		appUI.StopMonitoring()
		os.Exit(0)
	}()

	// Run application - window starts hidden (background mode)
	// User can show it with Alt+V or tray menu
	// X button hides window instead of quitting (tray menu has Quit option)
	appUI.Run()

	// Cleanup on normal exit
	hotkeyMgr.Stop()
	appUI.StopMonitoring()
}
