package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	cli "github.com/spf13/pflag"

	"github.com/MrZloHex/monolink"
	"monoview/internal/app"
)

const (
	NodeName = "MONOVIEW"
)

func main() {
	loadDotenv()

	defaultURLVal := envOr("MONOVIEW_URL", "wss://127.0.0.1:8443")
	defaultLogPath := envOr("MONOVIEW_LOG", "monoview.log")
	defaultTLSCert := os.Getenv("MONOVIEW_TLS_CERT")
	defaultTLSKey := os.Getenv("MONOVIEW_TLS_KEY")
	defaultTLSCA := os.Getenv("MONOVIEW_TLS_CA")
	defaultTLSServerName := os.Getenv("MONOVIEW_TLS_SERVER_NAME")

	url := cli.StringP("url", "u", defaultURLVal, "Url of hub (env MONOVIEW_URL)")
	tlsCert := cli.String("tls-cert", defaultTLSCert, "Client certificate PEM for mTLS (wss) (env MONOVIEW_TLS_CERT)")
	tlsKey := cli.String("tls-key", defaultTLSKey, "Client private key PEM for mTLS (wss) (env MONOVIEW_TLS_KEY)")
	tlsCA := cli.String("tls-ca", defaultTLSCA, "Optional CA PEM to verify server; default system roots (env MONOVIEW_TLS_CA)")
	tlsServerName := cli.String("tls-server-name", defaultTLSServerName, "TLS ServerName (SNI); use when URL is an IP (env MONOVIEW_TLS_SERVER_NAME)")
	logPath := cli.String("log-path", defaultLogPath, "Path to log file (env MONOVIEW_LOG)")
	cli.Parse()

	logFile, err := os.OpenFile(*logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot open log file %s: %v\n", *logPath, err)
		os.Exit(1)
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags)
	logger.Printf("monoview starting, url=%s", *url)

	var hubOpts []monolink.Option
	hubOpts = append(hubOpts,
		monolink.WithInbox(64),
		monolink.WithLogger(logger),
	)

	switch {
	case *tlsCert != "" && *tlsKey != "":
		cfg, err := monolink.LoadClientTLS(*tlsCert, *tlsKey, *tlsCA)
		if err != nil {
			fmt.Fprintf(os.Stderr, "mTLS: %v\n", err)
			os.Exit(1)
		}
		if *tlsServerName != "" {
			cfg.ServerName = *tlsServerName
		}
		hubOpts = append(hubOpts, monolink.WithTLS(cfg))
	case *tlsCert != "" || *tlsKey != "":
		fmt.Fprintln(os.Stderr, "mTLS requires both --tls-cert and --tls-key (or MONOVIEW_TLS_CERT and MONOVIEW_TLS_KEY)")
		os.Exit(1)
	}

	hub := monolink.New(NodeName, *url, hubOpts...)

	ctx := context.Background()
	if err := hub.Connect(ctx); err != nil {
		logger.Printf("concentrator offline: %v", err)
		hub = nil
	}

	m := app.NewModel()
	m.Hub = hub

	p := tea.NewProgram(m, tea.WithAltScreen())
	if hub != nil {
		if inbox := hub.Inbox(); inbox != nil {
			go func() {
				for {
					msg, ok := <-inbox
					if !ok {
						return
					}
					p.Send(app.HubMsg(msg))
				}
			}()
		}
	}
	if _, err := p.Run(); err != nil {
		logger.Printf("fatal: %v", err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if hub != nil {
		hub.Close()
	}
	logger.Printf("monoview stopped")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// loadDotenv runs before flags so MONOVIEW_* and MONO_ENV_FILE from .env apply everywhere.
// Path: MONO_ENV_FILE, else --env-file from argv, else ".env". Missing file is ignored.
func loadDotenv() {
	path := os.Getenv("MONO_ENV_FILE")
	if path == "" {
		path = dotenvPathFromArgs()
	}
	if path == "" {
		path = ".env"
	}
	err := godotenv.Load(path)
	if err == nil {
		return
	}
	if errors.Is(err, os.ErrNotExist) {
		return
	}
	var pe *os.PathError
	if errors.As(err, &pe) && errors.Is(pe.Err, os.ErrNotExist) {
		return
	}
	fmt.Fprintf(os.Stderr, "monoview: load %s: %v\n", path, err)
	os.Exit(1)
}

func dotenvPathFromArgs() string {
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--env-file" && i+1 < len(args):
			return args[i+1]
		case strings.HasPrefix(args[i], "--env-file="):
			return strings.TrimPrefix(args[i], "--env-file=")
		}
	}
	return ""
}
