package main

import (
	version "github.com/sniperkit/snk.golang.notify/pkg/version"
)

const (

	// programName sets...
	programName = `notify`

	// notifyVersion sets the notify library version used
	// nb. this is set with ldflags in Makefile...
	libNotifyVersion = version.Version

	// usage sets the template for the usage information message...
	usage = `usage: notify [-c command] [-f script file] [path]...

Listens on filesystem changes and forwards received mapping to
user-defined handlers.

The -c flag registers a command handler, which uses the syntax
of package template. Notify passes struct to the template,
splits produced string into command and args, and runs it using
exec.Command(). Additionaly the path and event type values are
accesible to the process via NOTIFY_PATH and NOTIFY_EVENT
environment variables.

The struct being passed to the template is:

	type Event struct {
		Path  string
		Event string
	}

Values for the Event field are:

	- create
	- remove
	- rename
	- write

The -t flag registers a file handler, which works similary
to the -c handler. The only difference the template is read from
the given file instead of the command line.

The path argument tells notify which director or directories to
listen on. By default notify listens recursively in current working
directory.

If no handler is specified notify prints each event to os.Stdout.`
)
