// Package tray implements methods that handle tray icon creation and removal
package tray

import (
	"github.com/getlantern/systray"
	"github.com/sht/lightfx/assets"
	"os"
	"os/signal"
	"syscall"
)

// init sets up interrupt handlers (Ctrl+C for example), and starts a goroutine
// that initializes the tray
func init() {
	handleInterrupts()
	go systray.Run(onReady, onExit)
}

func handleInterrupts() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		// will be reached when interrupted
		systray.Quit()
	}()
}

// onReady sets up the tray (label, icon, quit option etc.)
func onReady() {
	// Set icon
	ico := assets.GetIcon()
	systray.SetIcon(ico)
	// Setup menu items
	title := systray.AddMenuItem("Client", "")
	title.Disable()
	systray.AddSeparator()
	quit := systray.AddMenuItem("Quit", "")
	// Run an infinite loop on goroutine to detect button presses
	go func() {
		<-quit.ClickedCh
		systray.Quit()
	}()
}

// onExit is called when the tray icon is removed
func onExit() {
	// cleanup operations
	os.Exit(0)
}
