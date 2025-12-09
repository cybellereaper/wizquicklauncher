//go:build windows
// +build windows

package main

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// WindowManager handles all window-related operations
type WindowManager struct {
	user32         *windows.LazyDLL
	setWindowPos   *windows.LazyProc
	getClassNameW  *windows.LazyProc
	postMessageW   *windows.LazyProc
	setWindowTextW *windows.LazyProc
	enumWindows    *windows.LazyProc
	mu             sync.Mutex
}

// NewWindowManager creates a new window manager instance
func NewWindowManager() *WindowManager {
	user32 := windows.NewLazySystemDLL("user32.dll")
	return &WindowManager{
		user32:         user32,
		setWindowPos:   user32.NewProc("SetWindowPos"),
		getClassNameW:  user32.NewProc("GetClassNameW"),
		postMessageW:   user32.NewProc("PostMessageW"),
		setWindowTextW: user32.NewProc("SetWindowTextW"),
		enumWindows:    user32.NewProc("EnumWindows"),
	}
}

// MoveWindow moves the specified window to the given position
func (wm *WindowManager) MoveWindow(handle windows.Handle, x, y int) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.setWindowPos.Call(uintptr(handle), 0, uintptr(x), uintptr(y), 0, 0, 0x0001)
}

// GetAllWizardHandles returns a map of all Wizard101 window handles
func (wm *WindowManager) GetAllWizardHandles() map[windows.Handle]struct{} {
	handles := make(map[windows.Handle]struct{})
	callback := syscall.NewCallback(func(hwnd windows.Handle, _ uintptr) uintptr {
		buf := make([]uint16, 256)
		wm.getClassNameW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), 256)
		if windows.UTF16ToString(buf) == "Wizard Graphical Client" {
			handles[hwnd] = struct{}{}
		}
		return 1
	})
	wm.mu.Lock()
	wm.enumWindows.Call(callback, 0)
	wm.mu.Unlock()
	return handles
}

// SendChars sends characters to the specified window
func (wm *WindowManager) SendChars(handle windows.Handle, chars string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	for _, char := range chars {
		wm.postMessageW.Call(uintptr(handle), 0x102, uintptr(char), 0)
	}
}

// WizardLogin performs login in the Wizard101 window
func (wm *WindowManager) WizardLogin(handle windows.Handle, username, password string) {
	wm.SendChars(handle, username+"\t"+password+"\r")
	wm.mu.Lock()
	wm.setWindowTextW.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(fmt.Sprintf("[%s] Wizard101", username)))),
	)
	wm.mu.Unlock()
}
