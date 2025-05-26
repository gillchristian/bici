package main

import (
	"context"
	"fmt"

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
	events, err := calendar.ImportEvents(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return events
}

func (a *App) ExportEvents(events []calendar.CalendarEvent) string {
	err := calendar.ExportEvents(events)
	if err != nil {
		fmt.Println(err)
		return err.Error()
	}
	return ""
}
