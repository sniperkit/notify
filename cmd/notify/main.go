// Copyright (c) 2014-2015 The Notify Authors. All rights reserved.
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	// internal - core
	notify "github.com/sniperkit/snk.golang.notify/pkg"
)

var (
	file    string
	command string
	paths   = []string{"." + string(os.PathSeparator) + "..."}
	env     = newenv()
)

var mapping = map[notify.Event]string{
	notify.Create: "create",
	notify.Remove: "remove",
	notify.Rename: "rename",
	notify.Write:  "write",
}

func init() {
	flag.CommandLine.Usage = func() {
		fmt.Fprintln(os.Stderr, usage)
	}
	flag.StringVar(&file, "f", "", "script file to execute on received event")
	flag.StringVar(&command, "c", "", "command to run on received event")
	flag.Parse()
	if flag.NArg() != 0 {
		paths = flag.Args()
	}
}

func main() {
	var handlers []*handler
	if command != "" {
		h, err := newHandler(command)
		if err != nil {
			die(err)
		}
		handlers = append(handlers, h)
	}
	if file != "" {
		p, err := ioutil.ReadFile(file)
		if err != nil {
			die(err)
		}
		h, err := newHandler(string(p))
		if err != nil {
			die(err)
		}
		handlers = append(handlers, h)
	}
	var run []chan<- Event
	for _, h := range handlers {
		run = append(run, h.Daemon())
	}
	c := make(chan notify.EventInfo, 1)
	for _, path := range paths {
		if err := notify.Watch(path, c, notify.All); err != nil {
			die(err)
		}
	}
	for ei := range c {
		log.Println("received", ei)
		e := newEvent(ei)
		for _, run := range run {
			select {
			case run <- e:
			default:
				log.Println("event dropped due to slow handler")
			}
		}
	}
}

func die(v interface{}) {
	fmt.Fprintln(os.Stderr, v)
	os.Exit(1)
}
