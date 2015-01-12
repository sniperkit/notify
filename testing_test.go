package notify

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func isDir(path string) bool {
	r := path[len(path)-1]
	return r == '\\' || r == '/'
}

func tmpcreateall(tmp string, path string) error {
	isdir := isDir(path)
	path = filepath.Join(tmp, filepath.FromSlash(path))
	if isdir {
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	} else {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		if err := nonil(f.Sync(), f.Close()); err != nil {
			return err
		}
	}
	return nil
}

func tmpcreate(root, path string) (bool, error) {
	isdir := isDir(path)
	path = filepath.Join(root, filepath.FromSlash(path))
	if isdir {
		if err := os.Mkdir(path, 0755); err != nil {
			return false, err
		}
	} else {
		f, err := os.Create(path)
		if err != nil {
			return false, err
		}
		if err := nonil(f.Sync(), f.Close()); err != nil {
			return false, err
		}
	}
	return isdir, nil
}

// tmptree TODO
func tmptree(root, list string) (string, error) {
	f, err := os.Open(list)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if root == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		if root, err = ioutil.TempDir(pwd, "tmptree"); err != nil {
			return "", err
		}
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if err := tmpcreateall(root, scanner.Text()); err != nil {
			return "", err
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return root, nil
}

func TestDebug(t *testing.T) {
	var root string
	var args bool
	for _, arg := range os.Args {
		switch {
		case root != "":
			t.Skip("too many arguments")
		case args:
			root = arg
		case arg == "--":
			args = true
		}
	}
	if root == "" {
		t.Skip()
	}
	if _, err := tmptree(root, filepath.Join("testdata", "gopath.txt")); err != nil {
		t.Fatalf(`want tmptree(%q, "testdata/gopath.txt")=nil; got %v`, root, err)
	}
	fmt.Println(root)
}

// WCase TODO
type WCase struct {
	Action func()
	Events []EventInfo
}

// W TODO
type W struct {
	// Watcher TODO
	Watcher Watcher

	// C TODO
	C <-chan EventInfo

	// Timeout TODO
	Timeout time.Duration

	t     *testing.T
	root  string
	debug []string // NOTE for debugging only
}

func newWatcherTest(t *testing.T, tree string) *W {
	t.Parallel()
	root, err := tmptree("", filepath.FromSlash(tree))
	if err != nil {
		t.Fatalf(`tmptree("", %q)=%v`, tree, err)
	}
	Sync()
	return &W{
		t:    t,
		root: root,
	}
}

// NewWatcherTest TODO
func NewWatcherTest(t *testing.T, tree string, events ...Event) *W {
	w := newWatcherTest(t, tree)
	if len(events) == 0 {
		events = []Event{Create, Delete, Write, Move}
	}
	if rw, ok := w.watcher().(RecursiveWatcher); ok {
		if err := rw.RecursiveWatch(w.root, joinevents(events)); err != nil {
			t.Fatalf("RecursiveWatch(%q, All)=%v", w.root, err)
		}
	} else {
		fn := func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if fi.IsDir() {
				if err := w.watcher().Watch(path, joinevents(events)); err != nil {
					return err
				}
			}
			return nil
		}
		if err := filepath.Walk(w.root, fn); err != nil {
			t.Fatalf("Walk(%q, fn)=%v", w.root, err)
		}
	}
	return w
}

func (w *W) printdebug() {
	fmt.Println("[D] [WATCHER_TEST] to debugduce manually:")
	fmt.Printf("go test -run TestDebug -- %s\n", w.root)
	fmt.Println("cd", w.root)
	for _, debug := range w.debug {
		fmt.Println(debug)
	}
	fmt.Println()
}

// Fatal TODO
func (w *W) Fatal(v interface{}) {
	if dbg {
		w.printdebug()
	}
	w.t.Fatal(v)
}

// Fatalf TODO
func (w *W) Fatalf(format string, v ...interface{}) {
	if dbg {
		w.printdebug()
	}
	w.t.Fatalf(format, v...)
}

// Debug TODO
func (w *W) Debug(command string) {
	w.debug = append(w.debug, command)
}

func (w *W) initwatcher(buffer int) {
	c := make(chan EventInfo, buffer)
	w.Watcher = NewWatcher(c)
	w.C = c
}

func (w *W) watcher() Watcher {
	if w.Watcher == nil {
		w.initwatcher(512)
	}
	return w.Watcher
}

func (w *W) c() <-chan EventInfo {
	if w.C == nil {
		w.initwatcher(512)
	}
	return w.C
}

func (w *W) timeout() time.Duration {
	if w.Timeout != 0 {
		return w.Timeout
	}
	return 2 * time.Second
}

// Close TODO
func (w *W) Close() error {
	defer os.RemoveAll(w.root)
	// TODO(rjeczalik): make Close part of Watcher interface
	if err := w.watcher().(io.Closer).Close(); err != nil {
		w.Fatalf("w.Watcher.Close()=%v", err)
	}
	return nil
}

// Equal TODO
func EqualEventInfo(want, got EventInfo) error {
	if got.Event() != want.Event() {
		return fmt.Errorf("want ei.Event()=%v; got %v (path=%s)", want.Event(),
			got.Event(), want.Path())
	}
	if path := strings.TrimRight(want.Path(), `/\`); !strings.HasSuffix(got.Path(), path) {
		return fmt.Errorf("want ei.Path()=%s; got %s (event=%v)", want.Path(),
			got.Path(), want.Event())
	}
	return nil
}

// EqualCall TODO(rjeczalik)
func EqualCall(want, got Call) error {
	if got.E != want.E {
		return fmt.Errorf("want E=%v; got %v", want.E, got.E)
	}
	if got.NE != want.NE {
		return fmt.Errorf("want NE=%v; got %v", want.NE, got.NE)
	}
	if want.C != got.C {
		return fmt.Errorf("want C=%p; got %p", want.C, got.C)
	}
	if want := filepath.FromSlash(want.P); !strings.HasSuffix(got.P, want) {
		return fmt.Errorf("want P=%s; got %s", want, got.P)
	}
	if want := filepath.FromSlash(want.NP); !strings.HasSuffix(got.NP, want) {
		return fmt.Errorf("want NP=%s; got %s", want, got.NP)
	}
	if want.F != got.F {
		return fmt.Errorf("want F=%v; got %v", want.F, got.F)
	}
	return nil
}

// create TODO
func create(w *W, path string) WCase {
	return WCase{
		Action: func() {
			isdir, err := tmpcreate(w.root, filepath.FromSlash(path))
			if err != nil {
				w.Fatalf("tmpcreate(%q, %q)=%v", w.root, path, err)
			}
			if isdir {
				w.Debug(fmt.Sprintf("mkdir %s", path))
			} else {
				w.Debug(fmt.Sprintf("touch %s", path))
			}
		},
		Events: []EventInfo{
			&Call{P: path, E: Create},
		},
	}
}

// remove TODO
func remove(w *W, path string) WCase {
	return WCase{
		Action: func() {
			if err := os.RemoveAll(filepath.Join(w.root, filepath.FromSlash(path))); err != nil {
				w.Fatal(err)
			}
			w.Debug(fmt.Sprintf("rm -rf %s", path))
		},
		Events: []EventInfo{
			&Call{P: path, E: Delete},
		},
	}
}

// rename TODO
func rename(w *W, oldpath, newpath string) WCase {
	return WCase{
		Action: func() {
			err := os.Rename(filepath.Join(w.root, filepath.FromSlash(oldpath)),
				filepath.Join(w.root, filepath.FromSlash(newpath)))
			if err != nil {
				w.Fatal(err)
			}
			w.Debug(fmt.Sprintf("mv %s %s", oldpath, newpath))
		},
		Events: []EventInfo{
			&Call{P: newpath, E: Move},
		},
	}
}

// write TODO
func write(w *W, path string, p []byte) WCase {
	return WCase{
		Action: func() {
			f, err := os.OpenFile(filepath.Join(w.root, filepath.FromSlash(path)),
				os.O_WRONLY, 0644)
			if err != nil {
				w.Fatalf("OpenFile(%q)=%v", path, err)
			}
			if _, err := f.Write(p); err != nil {
				w.Fatalf("Write(%q)=%v", path, err)
			}
			if err := nonil(f.Sync(), f.Close()); err != nil {
				w.Fatalf("Sync(%q)/Close(%q)=%v", path, path, err)
			}
			w.Debug(fmt.Sprintf("echo %q > %s", p, path))
		},
		Events: []EventInfo{
			&Call{P: path, E: Write},
		},
	}
}

// ExpectAny TODO
func (w *W) ExpectAny(cases []WCase) {
Test:
	for i, cas := range cases {
		cas.Action()
		Sync()
		want := cas.Events[0]
		select {
		case ei := <-w.C:
			dbg.Printf("received: path=%q, event=%v, sys=%v (i=%d)", ei.Path(),
				ei.Event(), ei.Sys(), i)
			for j, want := range cas.Events {
				if err := EqualEventInfo(want, ei); err != nil {
					dbg.Print(err, j)
					continue
				}
				continue Test
			}
		case <-time.After(w.timeout()):
			w.Fatalf("timed out after %v waiting for %v (i=%d)", w.timeout(), want, i)
		}
	}
}

// FuncType represents enums for Watcher interface.
type FuncType string

const (
	FuncWatch            = FuncType("Watch")
	FuncUnwatch          = FuncType("Unwatch")
	FuncRewatch          = FuncType("Rewatch")
	FuncRecursiveWatch   = FuncType("RecursiveWatch")
	FuncRecursiveUnwatch = FuncType("RecursiveUnwatch")
	FuncRecursiveRewatch = FuncType("RecursiveRewatch")
	FuncStop             = FuncType("Stop")
)

// Chans TODO
type Chans []chan EventInfo

// Foreach TODO
func (c Chans) Foreach(fn func(chan<- EventInfo, Node)) {
	for i, ch := range c {
		fn(ch, Node{Name: strconv.Itoa(i)})
	}
}

// Drain TODO
func (c Chans) Drain() (ei []EventInfo) {
	n := len(c)
	stop := make(chan struct{})
	eich := make(chan EventInfo, n*buffer)
	go func() {
		defer close(eich)
		cases := make([]reflect.SelectCase, n+1)
		for i := range c {
			cases[i].Chan = reflect.ValueOf(c[i])
			cases[i].Dir = reflect.SelectRecv
		}
		cases[n].Chan = reflect.ValueOf(stop)
		cases[n].Dir = reflect.SelectRecv
		for {
			i, v, ok := reflect.Select(cases)
			if i == n {
				return
			}
			if !ok {
				panic("(Chans).Drain(): unexpected chan close")
			}
			eich <- v.Interface().(EventInfo)
		}
	}()
	<-time.After(timeout * time.Duration(n))
	close(stop)
	for e := range eich {
		ei = append(ei, e)
	}
	return
}

// NewChans TODO
func NewChans(n int) Chans {
	ch := make([]chan EventInfo, n)
	for i := range ch {
		ch[i] = make(chan EventInfo, buffer)
	}
	return ch
}

// Call represents single call to Watcher issued by the Tree
// and recorded by a spy Watcher mock.
type Call struct {
	F  FuncType       //
	C  chan EventInfo //
	P  string         // regular Path argument and old path from RecursiveRewatch call
	NP string         // new Path argument from RecursiveRewatch call
	E  Event          // regular Event argument and old Event from a Rewatch call
	NE Event          // new Event argument from Rewatch call
	S  interface{}    // when Call is used as EventInfo, S is a value of Sys()
}

// Call implements an EventInfo interface.
func (c *Call) Event() Event     { return c.E }
func (c *Call) Path() string     { return c.P }
func (c *Call) String() string   { return fmt.Sprintf("%#v", c) }
func (c *Call) Sys() interface{} { return c.S }

// Spy is a mock for Watcher interface, which records every call.
type Spy []Call

func (s Spy) Close() (_ error) { return }

func (s *Spy) Watch(p string, e Event) (_ error) {
	*s = append(*s, Call{F: FuncWatch, P: p, E: e})
	return
}

func (s *Spy) Unwatch(p string) (_ error) {
	*s = append(*s, Call{F: FuncUnwatch, P: p})
	return
}

func (s *Spy) Rewatch(p string, olde, newe Event) (_ error) {
	*s = append(*s, Call{F: FuncRewatch, P: p, E: olde, NE: newe})
	return
}

func (s *Spy) RecursiveWatch(p string, e Event) (_ error) {
	*s = append(*s, Call{F: FuncRecursiveWatch, P: p, E: e})
	return
}

func (s *Spy) RecursiveUnwatch(p string) (_ error) {
	*s = append(*s, Call{F: FuncRecursiveUnwatch, P: p})
	return
}

func (s *Spy) RecursiveRewatch(oldp, newp string, olde, newe Event) (_ error) {
	*s = append(*s, Call{F: FuncRecursiveRewatch, P: oldp, NP: newp, E: olde, NE: newe})
	return
}

// RCase TODO(rjeczalik)
type RCase struct {
	Call   Call
	Record []Call
}

// TCase TODO(rjeczalik)
type TCase struct {
	Event    Call
	Receiver Chans
}

// NCase TODO(rjeczalik)
type NCase struct {
	Action   WCase
	Receiver Chans
}

// N TODO(rjeczalik)
type N struct {
	// Notifier TODO(rjeczalik)
	Notifier Notifier

	t   *testing.T
	w   *W
	spy *Spy
	c   chan<- EventInfo
}

func newN(t *testing.T, tree string) *N {
	return &N{
		t: t,
		w: newWatcherTest(t, tree),
	}
}

func newTreeN(t *testing.T, tree string, fn func(spy *Spy) Watcher) *N {
	n := newN(t, tree)
	n.spy = &Spy{}
	c := make(chan EventInfo, 512)
	n.w.Watcher = fn(n.spy)
	n.w.C = c
	n.c = c
	n.Notifier = NewNotifier(n.w.watcher(), n.w.c())
	return n
}

// NewNotifyTest TODO(rjeczalik)
func NewNotifyTest(t *testing.T, tree string) *N {
	n := newN(t, tree)
	n.Notifier = NewNotifier(n.w.watcher(), n.w.c())
	return n
}

// NewTreeTest TODO(rjeczalik)
func NewTreeTest(t *testing.T, tree string) *N {
	fn := func(spy *Spy) Watcher {
		return struct {
			Watcher
		}{spy}
	}
	return newTreeN(t, tree, fn)
}

// NewRecursiveTreeTest TODO(rjeczalik)
func NewRecursiveTreeTest(t *testing.T, tree string) *N {
	fn := func(spy *Spy) Watcher { return spy }
	return newTreeN(t, tree, fn)
}

// Close TODO
func (n *N) Close() error {
	return n.w.Close()
}

// Watch TODO(rjeczalik)
func (n *N) Watch(path string, c chan<- EventInfo, events ...Event) error {
	return n.Notifier.Watch(filepath.Join(n.w.root, path), c, events...)
}

// Stop TODO(rjeczalik)
func (n *N) Stop(c chan<- EventInfo) {
	n.Notifier.Stop(c)
}

// Call TODO(rjeczalik)
func (n *N) Call(call Call) error {
	switch call.F {
	case FuncWatch:
		return n.Watch(call.P, call.C, call.E)
	case FuncStop:
		n.Stop(call.C)
		return nil
	default:
		return errors.New("unsupported call type: " + string(call.F))
	}
}

// ExpectRecordedCalls TODO(rjeczalik)
func (n *N) ExpectRecordedCalls(cases []RCase) {
	j := 0
	for i, cas := range cases {
		if err := n.Call(cas.Call); err != nil {
			n.t.Fatalf("Call(%v)=%v (i=%d)", cas, err, i)
			continue
		}
		record := (*n.spy)[j:]
		if len(cas.Record) == 0 && len(record) == 0 {
			continue
		}
		j = len(*n.spy)
		if len(record) != len(cas.Record) {
			n.t.Fatalf("want len(record)=%d; got %d (i=%d)", len(cas.Record),
				len(record), i)
		}
		for k := range cas.Record {
			if err := EqualCall(cas.Record[k], record[k]); err != nil {
				n.t.Fatal(err, i)
			}
		}
	}
}

// ExpectTreeEvents TODO(rjeczalik)
func (n *N) ExpectTreeEvents(cases []TCase) {
	for i, cas := range cases {
		cas.Event.P = filepath.Join(n.w.root, filepath.FromSlash(cas.Event.P))
		x := len(cas.Receiver)
		done := make(chan struct{})
		ev := make(map[<-chan EventInfo]EventInfo)
		go func() {
			cases := make([]reflect.SelectCase, x)
			for i := range cases {
				cases[i].Chan = reflect.ValueOf(cas.Receiver[i])
				cases[i].Dir = reflect.SelectRecv
			}
			for x := len(cases); x > 0; x = len(cases) {
				j, v, ok := reflect.Select(cases)
				if !ok {
					n.t.Fatalf("unexpected chan close (i=%d, j=%d)", i, j)
				}
				ch := cases[j].Chan.Interface().(chan EventInfo)
				ev[ch] = v.Interface().(EventInfo)
				cases[j], cases = cases[x-1], cases[:x-1]
			}
			close(done)
		}()
		n.c <- &cas.Event
		select {
		case <-done:
		case <-time.After(n.w.timeout()):
			n.t.Fatalf("ExpectTreeEvents has timed out after %v (i=%d)", n.w.timeout(), i)
		}
		for _, got := range ev {
			if err := EqualEventInfo(&cas.Event, got); err != nil {
				n.t.Error(err, i)
			}
		}
	}
}

// ExpectNotifyEvents TODO(rjeczalik)
func (n *N) ExpectNotifyEvents(cases []NCase) {
	// TODO
}