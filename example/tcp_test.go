package example_test

import (
	"crypto/tls"
	"os"
	"testing"

	"github.com/aacfactory/afssl"
)

const (
	addrName = `127.0.0.1:13000`
)

func TestTCP(t *testing.T) {
	certPEM, certErr := os.ReadFile(`/home/radxa/temp/poweroffs/ca.crt`)
	if certErr != nil {
		t.Error(certErr)
		return
	}
	keyPEM, keyErr := os.ReadFile(`/home/radxa/temp/poweroffs/ca.key`)
	if keyErr != nil {
		t.Error(keyErr)
		return
	}
	_, tlsConfig, sccErr := afssl.SSC(certPEM, keyPEM)
	if sccErr != nil {
		t.Error(sccErr)
		return
	}

	conn, connErr := tls.Dial("tcp", addrName, tlsConfig)
	if connErr != nil {
		t.Error(connErr)
		return
	}

	_, wErr := conn.Write([]byte("reboot"))
	if wErr != nil {
		t.Error(wErr)
		return
	}
	_ = conn.Close()
}
