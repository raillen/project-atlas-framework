package indexing

import "testing"

func TestChanged(t *testing.T) {
	p := Index{Entries: []Entry{{Path: "a", Hash: "1"}}}
	c := Index{Entries: []Entry{{Path: "a", Hash: "2"}, {Path: "b", Hash: "1"}}}
	got := Changed(p, c)
	if len(got) != 2 {
		t.Fatal(got)
	}
}
