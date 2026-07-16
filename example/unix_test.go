package example_test

import (
	"net"
	"os"
	"path/filepath"
	"testing"
)

const (
	addrFilename = `temp/poweroffs/poweroffs.sock`
)

func TestUNIX(t *testing.T) {
	home, homeErr := os.UserHomeDir()
	if homeErr != nil {
		t.Error(homeErr)
		return
	}
	addr := filepath.Join(home, addrFilename)

	conn, connErr := net.Dial("unix", addr)
	if connErr != nil {
		t.Error(connErr)
		return
	}

	_, wErr := conn.Write([]byte("poweroff"))
	if wErr != nil {
		t.Error(wErr)
		return
	}
	_ = conn.Close()
}
