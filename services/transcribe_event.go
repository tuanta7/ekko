package services

func (t *TranscribeService) emit(name string, data any) {
	t.mu.Lock()
	app := t.app // Lock briefly to safely read the shared app
	t.mu.Unlock()

	if app != nil {
		app.Event.Emit(name, data)
	}
}
