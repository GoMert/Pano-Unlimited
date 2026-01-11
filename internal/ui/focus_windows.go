//go:build windows
// +build windows

package ui

import (
	"syscall"
	"unsafe"
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procShowWindow          = user32.NewProc("ShowWindow")
	procFindWindowW         = user32.NewProc("FindWindowW")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procAttachThreadInput   = user32.NewProc("AttachThreadInput")
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procGetCurrentThreadId  = kernel32.NewProc("GetCurrentThreadId")
)

const (
	SW_SHOW    = 5
	SW_RESTORE = 9
)

// BringWindowToFront forcefully brings a window to the foreground on Windows
func BringWindowToFront(windowTitle string) {
	// Convert window title to UTF16
	titlePtr, _ := syscall.UTF16PtrFromString(windowTitle)
	
	// Find window by title
	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	if hwnd == 0 {
		return
	}

	// Get foreground window
	foregroundHwnd, _, _ := procGetForegroundWindow.Call()
	
	// Get thread IDs
	var foregroundThreadId uint32
	procGetWindowThreadProcessId.Call(foregroundHwnd, uintptr(unsafe.Pointer(&foregroundThreadId)))
	
	currentThreadId, _, _ := procGetCurrentThreadId.Call()
	
	// Attach input threads to allow SetForegroundWindow
	if foregroundThreadId != uint32(currentThreadId) {
		procAttachThreadInput.Call(currentThreadId, uintptr(foregroundThreadId), 1)
		defer procAttachThreadInput.Call(currentThreadId, uintptr(foregroundThreadId), 0)
	}
	
	// Show and restore window if minimized
	procShowWindow.Call(hwnd, SW_RESTORE)
	
	// Bring to foreground
	procSetForegroundWindow.Call(hwnd)
}
