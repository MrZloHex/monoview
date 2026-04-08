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

	"monoview/internal/app"
	"monoview/pkg/concentrator"
)

const (
	defaultNode    = "MONOVIEW"
	defaultURL     = "ws://192.168.0.69:8092"
	defaultLogFile = "monoview.log"
)

func main() {
	loadDotenv()

	node := envOr("MONO_NODE", defaultNode)
	url := cli.StringP("url", "u", "ws://192.168.0.69:8092", "Url of hub; env MONO_URL after .env load")
	tlsCert := cli.String("tls-cert", "", "Client certificate PEM for mTLS (wss); env MONO_TLS_CERT")
	tlsKey := cli.String("tls-key", "", "Client private key PEM for mTLS; env MONO_TLS_KEY")
	tlsCA := cli.String("tls-ca", "", "Optional CA PEM to verify server; default system roots; env MONO_TLS_CA")
	tlsServerName := cli.String("tls-server-name", "", "TLS ServerName (SNI); use when URL is an IP; env MONO_TLS_SERVER_NAME")
	_ = cli.String("env-file", ".env", "Dotenv path (loaded before Parse); env MONO_ENV_FILE overrides")
	logPath := envOr("MONO_LOG", defaultLogFile)
	cli.Parse()

	hubURL := *url
	if u := cli.Lookup("url"); u != nil && !u.Changed {
		hubURL = envOr("MONO_URL", hubURL)
	}

	certPath := *tlsCert
	if c := cli.Lookup("tls-cert"); c != nil && !c.Changed {
		certPath = envOr("MONO_TLS_CERT", certPath)
	}
	keyPath := *tlsKey
	if c := cli.Lookup("tls-key"); c != nil && !c.Changed {
		keyPath = envOr("MONO_TLS_KEY", keyPath)
	}
	caPath := *tlsCA
	if c := cli.Lookup("tls-ca"); c != nil && !c.Changed {
		caPath = envOr("MONO_TLS_CA", caPath)
	}
	serverName := *tlsServerName
	if c := cli.Lookup("tls-server-name"); c != nil && !c.Changed {
		serverName = envOr("MONO_TLS_SERVER_NAME", serverName)
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot open log file %s: %v\n", logPath, err)
		os.Exit(1)
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags)
	logger.Printf("monoview starting, node=%s url=%s", node, hubURL)

	var hubOpts []concentrator.Option
	hubOpts = append(hubOpts,
		concentrator.WithInbox(64),
		concentrator.WithLogger(logger),
	)

	switch {
	case certPath != "" && keyPath != "":
		cfg, err := concentrator.LoadClientTLS(certPath, keyPath, caPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "mTLS: %v\n", err)
			os.Exit(1)
		}
		if serverName != "" {
			cfg.ServerName = serverName
		}
		hubOpts = append(hubOpts, concentrator.WithTLSConfig(cfg))
	case certPath != "" || keyPath != "":
		fmt.Fprintln(os.Stderr, "mTLS requires both --tls-cert and --tls-key (or MONO_TLS_CERT and MONO_TLS_KEY)")
		os.Exit(1)
	}

	hub := concentrator.New(node, hubURL, hubOpts...)

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

// loadDotenv runs before flags so MONO_* and TLS paths from .env apply everywhere.
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
