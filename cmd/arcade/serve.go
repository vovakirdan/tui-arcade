package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/vovakirdan/tui-arcade/internal/platform/tui"
)

var (
	flagSSHAddr     string
	flagHostKey     string
	flagSSHDBPath   string
	flagIdleTimeout int
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the arcade SSH server",
	Long: `Start an SSH server that allows users to connect and play games.

Each SSH connection gets their own session with a game picker menu.
Scores are stored per-server (all users share the same leaderboard).

Host key handling:
  - If --host-key is provided, uses that key file
  - Otherwise, auto-generates a key at ~/.arcade/host_key

Examples:
  arcade serve                           # Listen on :23234 with auto-generated key
  arcade serve --ssh :2222               # Listen on port 2222
  arcade serve --host-key ./my_host_key  # Use specific host key
  arcade serve --db ./scores.db          # Use specific database

Users can connect with:
  ssh localhost -p 23234`,
	Run: runServe,
}

func init() {
	serveCmd.Flags().StringVar(&flagSSHAddr, "ssh", ":23234", "SSH server address (host:port)")
	serveCmd.Flags().StringVar(&flagHostKey, "host-key", "", "Path to host key file (auto-generated if not specified)")
	serveCmd.Flags().StringVar(&flagSSHDBPath, "db", "~/.arcade/scores.db", "Path to scores database")
	serveCmd.Flags().IntVar(&flagIdleTimeout, "idle-timeout", 30, "Idle timeout in minutes before disconnecting")
}

func runServe(_ *cobra.Command, _ []string) {
	cfg := tui.SSHServerConfig{
		Address:     flagSSHAddr,
		HostKeyPath: flagHostKey,
		DBPath:      flagSSHDBPath,
		IdleTimeout: time.Duration(flagIdleTimeout) * time.Minute,
	}

	server, err := tui.NewSSHServer(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating server: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting arcade SSH server on %s\n", cfg.Address)
	fmt.Println("Connect with: ssh localhost -p 23234")
	fmt.Println("Press Ctrl+C to stop")

	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
