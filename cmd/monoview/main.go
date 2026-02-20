package main

import (
	"context"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"monoview/internal/ui"
	"monoview/pkg/concentrator"
)

const (
	defaultNode    = "MONOVIEW"
	defaultURL     = "ws://192.168.0.69:8092"
	defaultLogFile = "monoview.log"
)

func main() {
	node := envOr("MONO_NODE", defaultNode)
	url := envOr("MONO_URL", defaultURL)
	logPath := envOr("MONO_LOG", defaultLogFile)

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot open log file %s: %v\n", logPath, err)
		os.Exit(1)
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags)
	logger.Printf("monoview starting, node=%s url=%s", node, url)

	hub := concentrator.New(node, url,
		concentrator.WithInbox(64),
		concentrator.WithLogger(logger),
	)

	ctx := context.Background()
	if err := hub.Connect(ctx); err != nil {
		logger.Printf("concentrator offline: %v", err)
		hub = nil
	}

	m := ui.NewModel()
	m.Hub = hub

	p := tea.NewProgram(m, tea.WithAltScreen())
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
