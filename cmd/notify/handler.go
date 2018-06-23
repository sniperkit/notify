package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"text/template"

	// internal - plugin
	"github.com/sniperkit/snk.golang.notify/plugin/cli"
)

type handler struct {
	tmpl *template.Template
	env  []string
}

func newHandler(text string) (*handler, error) {
	tmpl, err := template.New("main.Handler").Parse(text)
	if err != nil {
		return nil, err
	}
	h := &handler{
		tmpl: tmpl,
		env:  env(Event{}),
	}
	return h, nil
}

func (h *handler) Run(e Event) error {
	var buf bytes.Buffer
	if err := h.tmpl.Execute(&buf, e); err != nil {
		return err
	}
	name, args := cli.Split(buf.String())
	cmd := exec.Command(name, args...)
	h.env[len(h.env)-1] = "NOTIFY_EVENT=" + e.Event
	h.env[len(h.env)-2] = "NOTIFY_PATH=" + e.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = h.env
	return cmd.Run()
}

func (h *handler) Daemon() chan<- Event {
	c := make(chan Event)
	go func() {
		for e := range c {
			if err := h.Run(e); err != nil {
				log.Println("handler error:", err)
			}
		}
	}()
	return c
}
