package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aacfactory/afssl"
)

var (
	ErrReboot   = errors.New("reboot")
	ErrPowerOff = errors.New("power-off")
)

const (
	reboot   = "reboot"
	poweroff = "poweroff"
)

type Options struct {
	UnixAddr     string
	TcpAddr      string
	CertFilename string
	KeyFilename  string
}

func Run(ctx context.Context, options Options) (err error) {
	var (
		unixListener net.Listener
		tcpListener  net.Listener
	)

	unixAddr := strings.TrimSpace(options.UnixAddr)
	if unixAddr == "" {
		err = errors.Join(errors.New("run failed"), errors.New("unix addr is required"))
		return
	}
	if !filepath.IsAbs(unixAddr) {
		abs, absErr := filepath.Abs(unixAddr)
		if absErr != nil {
			err = errors.Join(errors.New("run failed"), errors.New("could not resolve unix socket path"), absErr)
			return
		}
		unixAddr = abs
	}
	unixAddr = filepath.ToSlash(unixAddr)
	_ = os.Remove(unixAddr)

	if tcpAddr := strings.TrimSpace(options.TcpAddr); tcpAddr != "" {
		certFilename := strings.TrimSpace(options.CertFilename)
		keyFilename := strings.TrimSpace(options.KeyFilename)
		if certFilename == "" || keyFilename == "" {
			err = errors.Join(errors.New("run failed"), errors.New("enable tcp failed"), errors.New("tcp require ca file"))
			return
		}
		if !filepath.IsAbs(certFilename) {
			abs, absErr := filepath.Abs(certFilename)
			if absErr != nil {
				err = errors.Join(errors.New("run failed"), errors.New("enable tcp failed"), errors.New("could not resolve ca cert path"), absErr)
				return
			}
			certFilename = abs
		}
		if !filepath.IsAbs(keyFilename) {
			abs, absErr := filepath.Abs(keyFilename)
			if absErr != nil {
				err = errors.Join(errors.New("run failed"), errors.New("enable tcp failed"), errors.New("could not resolve ca key path"), absErr)
				return
			}
			keyFilename = abs
		}
		certFilename = filepath.ToSlash(certFilename)
		keyFilename = filepath.ToSlash(keyFilename)
		caPEM, caPEMErr := os.ReadFile(certFilename)
		if caPEMErr != nil {
			err = errors.Join(errors.New("run failed"), errors.New("enable tcp failed"), errors.New("could not read ca cert file"), caPEMErr)
			return
		}
		caKeyPEM, caKeyPEMErr := os.ReadFile(keyFilename)
		if caKeyPEMErr != nil {
			err = errors.Join(errors.New("run failed"), errors.New("enable tcp failed"), errors.New("could not read ca key file"), caKeyPEMErr)
			return
		}
		serverPEM, serverKeyPEM, serverErr := afssl.GenerateCertificate(afssl.CertificateConfig{}, afssl.WithParent(caPEM, caKeyPEM))
		if serverErr != nil {
			err = errors.Join(errors.New("run failed"), errors.New("enable tcp failed"), errors.New("could not generate server tls keypair from ca"), serverErr)
			return
		}
		cas := x509.NewCertPool()
		if !cas.AppendCertsFromPEM(caPEM) {
			err = errors.Join(errors.New("run failed"), errors.New("enable tcp failed"), errors.New("could append ca into cert pool"))
			return
		}
		serverCertificate, serverCertificateErr := tls.X509KeyPair(serverPEM, serverKeyPEM)
		if serverCertificateErr != nil {
			err = serverCertificateErr
			err = errors.Join(errors.New("run failed"), errors.New("enable tcp failed"), errors.New("could load server keypair"), serverCertificateErr)
			return
		}
		tlsConfig := &tls.Config{
			ClientCAs:    cas,
			Certificates: []tls.Certificate{serverCertificate},
			ClientAuth:   tls.RequireAndVerifyClientCert,
		}

		tcpListener, err = tls.Listen("tcp", tcpAddr, tlsConfig)
		if err != nil {
			return
		}
	}

	unixListener, err = net.Listen("unix", unixAddr)
	if err != nil {
		if tcpListener != nil {
			_ = tcpListener.Close()
		}
		err = errors.Join(errors.New("run failed"), errors.New("could not listen unix socket"), err)
		return
	}

	rch := make(chan string, 8)
	wg := new(sync.WaitGroup)

	wg.Add(1)
	go handle(unixListener, rch, wg)
	if tcpListener != nil {
		wg.Add(1)
		go handle(tcpListener, rch, wg)
	}

	signals := make([]string, 0, 2)
	select {
	case <-ctx.Done():
		err = ctx.Err()
		break
	case r := <-rch:
		signals = append(signals, r)
		break
	}

	_ = unixListener.Close()
	if tcpListener != nil {
		_ = tcpListener.Close()
	}
	wg.Wait()

	if err != nil {
		return
	}

	timer := time.NewTimer(500 * time.Millisecond)
	stopped := false
	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			stopped = true
			break
		case <-timer.C:
			stopped = true
			break
		case r := <-rch:
			signals = append(signals, r)
			break
		}
		if stopped {
			break
		}
	}
	timer.Stop()

	if err != nil {
		return
	}

	hasPowerOff := false
	for _, signal := range signals {
		if signal == poweroff {
			hasPowerOff = true
			break
		}
	}
	if hasPowerOff {
		err = ErrPowerOff
	} else {
		err = ErrReboot
	}

	return
}

func handle(ln net.Listener, rch chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		conn, err := ln.Accept()
		if err != nil {
			break
		}
		wg.Add(1)
		go func(conn net.Conn, rch chan<- string, wg *sync.WaitGroup) {
			defer wg.Done()

			buf := bytes.NewBuffer(nil)
			b := make([]byte, 8)
			_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
			cmd := ""
			for {
				rn, rErr := conn.Read(b)
				if rn > 0 {
					buf.Write(b[:rn])
					s := strings.ToLower(strings.TrimSpace(buf.String()))
					if s == reboot {
						cmd = reboot
						break
					} else if s == poweroff {
						cmd = poweroff
						break
					} else if len(s) >= max(len(reboot), len(poweroff)) {
						break
					}
				}
				if rErr != nil {
					break
				}
			}
			_ = conn.Close()
			buf.Reset()

			if cmd == reboot || cmd == poweroff {
				rch <- cmd
			}

		}(conn, rch, wg)
	}

}
