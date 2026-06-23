package services

func (ts *TranscribeService) emit(name string, data any) {
	ts.mu.Lock()
	app := ts.app // Lock briefly to safely read the shared app
	ts.mu.Unlock()

	if app != nil {
		app.Event.Emit(name, data)
	}
}
