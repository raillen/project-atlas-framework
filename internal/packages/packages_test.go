package packages

import "testing"

func TestLockSorts(t *testing.T) {
	lock := NewLock([]Resolved{{ID: "z"}, {ID: "a"}})
	if lock.Packages[0].ID != "a" {
		t.Fatal(lock)
	}
}
func TestChecksum(t *testing.T) {
	if len(Checksum([]byte("x"))) != 64 {
		t.Fatal("invalid checksum")
	}
}
