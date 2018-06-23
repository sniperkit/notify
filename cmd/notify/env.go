package main

import (
	"os"
	"strings"
)

func newenv() func(Event) []string {
	env := os.Environ()
	for i, s := range env {
		s = strings.ToLower(s)
		if strings.Contains(s, "NOTIFY_PATH=") || strings.Contains(s, "NOTIFY_EVENT=") {
			env[i], env = env[len(env)-1], env[:len(env)-1]
		}
	}
	env = append(env, "", "")
	return func(e Event) []string {
		s := make([]string, len(env))
		copy(s, env)
		s[len(s)-1] = "NOTIFY_EVENT=" + e.Event
		s[len(s)-2] = "NOTIFY_PATH=" + e.Path
		return s
	}
}
