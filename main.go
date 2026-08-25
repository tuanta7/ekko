package main

import (
	"embed"
	_ "embed"
	"log"

	"github.com/tuanta7/ekko/services"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var trayIcon []byte

func init() {

	// Transcribe events
	application.RegisterEvent[services.StateEvent]("transcribe:state")
	application.RegisterEvent[services.TranscriptEvent]("transcribe:partial")
	application.RegisterEvent[services.TranscriptEvent]("transcribe:final")
	application.RegisterEvent[services.ErrorEvent]("transcribe:error")
}

// The main function serves as the application's entry point. It initializes the application, creates a window,
// and starts a goroutine that emits a time-based event every second. It subsequently runs the application and
// logs any error that might occur.
func main() {
	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Bind' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running an macOS.
	app := application.New(application.Options{
		Name:        "Ekko",
		Description: "Local audio transcriber",
		Services: []application.Service{
			application.NewService(&services.TranscribeService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	// GTK4 windows are painted opaque by the theme, so drop that background once
	// GTK is up (on the main thread) or the transparent webview shows grey.
	app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
		application.InvokeSync(transparentWindows)
	})

	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Ekko",
		Frameless:        true,
		AlwaysOnTop:      true,
		Width:            480,
		MinWidth:         400,
		Height:           200,
		MinHeight:        150,
		BackgroundType:   application.BackgroundTypeTransparent,
		BackgroundColour: application.NewRGBA(0, 0, 0, 0), // fully transparent webview; the page's CSS paints the window
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			DisableShadow:           true,
			Backdrop:                application.MacBackdropTransparent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		URL: "/",
	})

	menu := application.NewMenu()
	menu.Add("Hide").OnClick(func(*application.Context) { window.Hide() })
	menu.Add("Show").OnClick(func(*application.Context) { window.Show() })
	menu.AddSeparator()
	menu.Add("Quit").OnClick(func(*application.Context) { app.Quit() })

	tray := app.SystemTray.New().SetIcon(trayIcon).SetMenu(menu)
	tray.OnClick(tray.ShowMenu)

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
