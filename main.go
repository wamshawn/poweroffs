package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/urfave/cli/v3"
	"github.com/wamshawn/poweroffs/service"
	"golang.org/x/sys/unix"
)

const (
	unixSocketParam       = "unix"
	defaultUnixSocketAddr = `/var/local/poweroffs/poweroffs.sock`
	tcpSocketParam        = "tcp"
	defaultTcpSocketAddr  = "0.0.0.0:13000"
	certParam             = "cert"
	keyParam              = "key"
)

const (
	generateOutputDir = "out"
)

const (
	logFilename = `/var/log/poweroffs.log`
)

func main() {
	// ctx
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGABRT, syscall.SIGKILL, syscall.SIGTERM)
	defer cancel()

	// cmd
	cmd := &cli.Command{
		Name:                  "poweroffs",
		Version:               "v0.0.1",
		EnableShellCompletion: true,
		Usage:                 "run a unix socket to power off system",
		Copyright:             "(c) 2026 CNXGM Enterprise",
		Commands: []*cli.Command{
			{
				Name:  "gen",
				Usage: "generate a certificate",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    generateOutputDir,
						Aliases: []string{"o"},
						Usage:   "output directory",
					},
				},
				Action: func(ctx context.Context, command *cli.Command) (err error) {
					output := strings.TrimSpace(command.String(generateOutputDir))

					if err = service.Gen(output); err != nil {
						return
					}

					fmt.Println("certificate ca generated successfully!")
					return
				},
			},
			{
				Name:  "run",
				Usage: "run a power off signal system",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    unixSocketParam,
						Value:   defaultUnixSocketAddr,
						Usage:   "unix sock file path",
						Sources: cli.EnvVars("POWEROFFS_UNIX_SOCK"),
					},
					&cli.StringFlag{
						Name:    tcpSocketParam,
						Value:   defaultTcpSocketAddr,
						Usage:   "tcp sock file path",
						Sources: cli.EnvVars("POWEROFFS_TCP_SOCK"),
					},
					&cli.StringFlag{
						Name:    certParam,
						Value:   "",
						Usage:   "ca cert file path",
						Sources: cli.EnvVars("POWEROFFS_CERT"),
					},
					&cli.StringFlag{
						Name:    keyParam,
						Value:   "",
						Usage:   "ca key file path",
						Sources: cli.EnvVars("POWEROFFS_KEY"),
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) (err error) {
					unixAddr := strings.TrimSpace(cmd.String(unixSocketParam))
					if unixAddr == "" {
						unixAddr = defaultUnixSocketAddr
					}
					tcpAddr := strings.TrimSpace(cmd.String(tcpSocketParam))
					certFilename := strings.TrimSpace(cmd.String(certParam))
					keyFilename := strings.TrimSpace(cmd.String(keyParam))

					err = service.Run(ctx, service.Options{
						UnixAddr:     unixAddr,
						TcpAddr:      tcpAddr,
						CertFilename: certFilename,
						KeyFilename:  keyFilename,
					})

					return
				},
			},
		},
	}

	// run
	args := os.Args

RERUN:
	rebootCMD := 0
	if err := cmd.Run(ctx, args); err != nil {
		if errors.Is(err, service.ErrReboot) {
			rebootCMD = unix.LINUX_REBOOT_CMD_RESTART
			s := fmt.Sprintf("[%s] poweroffs recv restart signal!!!\n", time.Now().Format(time.RFC3339))
			_ = os.WriteFile(logFilename, []byte(s), 0644)
			fmt.Println(s)
		} else if errors.Is(err, service.ErrPowerOff) {
			rebootCMD = unix.LINUX_REBOOT_CMD_POWER_OFF
			s := fmt.Sprintf("[%s] poweroffs recv power-off signal!!!\n", time.Now().Format(time.RFC3339))
			_ = os.WriteFile(logFilename, []byte(s), 0644)
			fmt.Println(s)
		} else if !errors.Is(err, context.Canceled) {
			buf := bytes.NewBuffer(nil)
			buf.WriteString(fmt.Sprintf("[%s] poweroffs run failed!!!\n", time.Now().Format(time.RFC3339)))
			buf.WriteString(fmt.Sprintf("%+v\n", err))
			_ = os.WriteFile(logFilename, buf.Bytes(), 0644)
			fmt.Println(buf.String())
			buf.Reset()
		}
	}
	if rebootCMD == 0 {
		return
	}

	time.Sleep(500 * time.Millisecond)
	unix.Sync()
	if err := unix.Reboot(rebootCMD); err != nil {
		buf := bytes.NewBuffer(nil)
		buf.WriteString(fmt.Sprintf("[%s] poweroffs reboot failed!!!\n", time.Now().Format(time.RFC3339)))
		buf.WriteString(fmt.Sprintf("%+v\n", err))
		_ = os.WriteFile(logFilename, buf.Bytes(), 0644)
		fmt.Println(buf.String())
		buf.Reset()
		goto RERUN
	}

}
