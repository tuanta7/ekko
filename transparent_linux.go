package main

/*
#cgo linux pkg-config: gtk4
#include <gtk/gtk.h>

// GTK4 toplevels carry the ".background" style class, and the theme paints it
// opaque (Default/Yaru: background-color: #f6f5f4). Wails' GTK4 backend only
// sets the *webview* background from BackgroundColour, so a transparent webview
// just reveals that opaque window behind it. Override it at APPLICATION
// priority, which outranks the theme.
static void ekko_transparent_windows(void) {
	GdkDisplay *display = gdk_display_get_default();
	if (display == NULL) {
		return;
	}
	GtkCssProvider *provider = gtk_css_provider_new();
	gtk_css_provider_load_from_string(provider, "window.background { background-color: transparent; }");
	gtk_style_context_add_provider_for_display(display, GTK_STYLE_PROVIDER(provider), GTK_STYLE_PROVIDER_PRIORITY_APPLICATION);
	g_object_unref(provider);
}
*/
import "C"

func transparentWindows() {
	C.ekko_transparent_windows()
}
