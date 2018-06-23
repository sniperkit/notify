package main

import (
	notify "github.com/sniperkit/snk.golang.notify/pkg"
)

type Event struct {
	Path  string
	Event string
}

// newEvent TODO(rjeczalik)
func newEvent(ei notify.EventInfo) Event {
	return Event{
		Path:  ei.Path(),
		Event: mapping[ei.Event()],
	}
}
