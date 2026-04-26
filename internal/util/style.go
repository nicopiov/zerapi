package util

import "github.com/fatih/color"

var (
	Success = color.New(color.FgGreen).SprintFunc()
	Error   = color.New(color.FgRed).SprintFunc()
	Warn    = color.New(color.FgYellow).SprintFunc()
	Info    = color.New(color.FgCyan).SprintFunc()
	Muted   = color.New(color.FgHiBlack).SprintFunc()
)

func Status(status int) string {
	switch {
	case status >= 200 && status < 300:
		return Success(status)
	case status >= 400:
		return Error(status)
	case status >= 300:
		return Warn(status)
	default:
		return Muted(status)
	}
}
