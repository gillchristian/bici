package main

import (
	"context"

	calendar "bici/calendar"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) ImportEvents(path string) []calendar.CalendarEvent {
	return calendar.ImportEvents(path)
}

func (a *App) ExportEvents(events []calendar.CalendarEvent) (bool, string) {
	return calendar.ExportEvents(events)
}
