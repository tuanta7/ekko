package console

type sourcesLoadedMsg struct {
	sources []string
	err     error
}

type sessionEndMsg struct {
	err error
}

type transcriptUpdateMsg struct {
	text string
}

type recordingErrorMsg struct {
	err error
}
