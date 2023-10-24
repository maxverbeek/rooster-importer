package domain

import (
	"io"
)

type Application struct {
	xlsxfile io.ReadCloser

	guistuff chan interface{}
}

func NewApplication() *Application {
	return &Application{
		guistuff: make(chan interface{}),
	}
}

func (a *Application) SubscribeToGui(events <-chan Action) {
	for action := range events {
		action(a)
	}
}

func (a *Application) GuiStuff() <-chan interface{} {
	return a.guistuff
}
