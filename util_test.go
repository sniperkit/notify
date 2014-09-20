package notify

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestAppendset(t *testing.T) {
	cases := [...]struct {
		start []string
		vals  map[string]struct{}
		end   []string
	}{
		0: {
			[]string{"a", "e", "x"},
			map[string]struct{}{"A": {}, "E": {}, "x": {}},
			[]string{"A", "E", "a", "e", "x"},
		},
		1: {
			[]string{},
			map[string]struct{}{"/a/b/c/d": {}, "/b/c/d": {}, "/c/d": {}, "/d": {}},
			[]string{"/a/b/c/d", "/b/c/d", "/c/d", "/d"},
		},
		2: {
			[]string{"a", "b", "c"},
			map[string]struct{}{"d": {}, "e": {}, "f": {}},
			[]string{"a", "b", "c", "d", "e", "f"},
		},
	}
	for i := range cases {
		s := cases[i].start
		for v := range cases[i].vals {
			s = appendset(s, v)
		}
		if !reflect.DeepEqual(s, cases[i].end) {
			t.Errorf("want s=%v; got %v (i=%d)", cases[i].end, s, i)
		}
	}
}

func TestSplitpath(t *testing.T) {
	cases := map[string][]string{
		"C:/a/b/c/d.txt": {"a", "b", "c", "d.txt"},
		"/a/b/c/d.txt":   {"a", "b", "c", "d.txt"},
		"":               nil,
		".":              nil,
		"C:":             nil,
	}
	for path, names := range cases {
		path = filepath.FromSlash(path)
		if s := splitpath(path); !reflect.DeepEqual(s, names) {
			t.Errorf("want s=%v; got %v (path=%s)", names, s, path)
		}
	}
}

func TestJoinevents(t *testing.T) {
	cases := [...]struct {
		evs []Event
		ev  Event
	}{
		0: {nil, All},
		1: {[]Event{}, All},
		2: {[]Event{Create}, Create},
		3: {[]Event{Move}, Move},
		4: {[]Event{Create, Write, Delete}, Create | Write | Delete},
	}
	for i, cas := range cases {
		if ev := joinevents(cas.evs); ev != cas.ev {
			t.Errorf("want event=%v; got %v (i=%d)", cas.ev, ev, i)
		}
	}
}

func TestSplitevents(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestWalkpath(t *testing.T) {
	cases := map[string]struct {
		p  []string
		ok bool
	}{
		"C:/a/b/c/d.txt":       {[]string{"a", "b", "c", "d.txt"}, true},
		"/a/b/c/d.txt":         {[]string{"a", "b", "c", "d.txt"}, true},
		"C:/a/b/c/break":       {[]string{"a", "b", "c", "break"}, true},
		"/a/b/c/break/":        {[]string{"a", "b", "c", "break"}, true},
		"C:/a/b/c/break/d.txt": {[]string{"a", "b", "c", "break"}, false},
		"/a/b/c/break/d.txt":   {[]string{"a", "b", "c", "break"}, false},
		"":                     {nil, false},
		".":                    {nil, false},
		"C:":                   {nil, false},
	}
	var p []string
	fn := func(s string) bool {
		p = append(p, s)
		return s != "break"
	}
	for path, cas := range cases {
		p, path = p[:0], filepath.FromSlash(path)
		if ok := walkpath(path, fn); ok != cas.ok {
			t.Errorf("want ok=%v; got %v (path=%s)", cas.ok, ok, path)
			continue
		}
		// Because reflect.DeepEqual([]string(nil), []string{}) gives false.
		if len(cas.p)+len(p) == 0 {
			continue
		}
		if !reflect.DeepEqual(p, cas.p) {
			t.Errorf("want p=%v; got %v (path=%s)", cas.p, p, path)
		}
	}
}
