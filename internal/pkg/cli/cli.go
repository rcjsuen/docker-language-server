package cli

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"

	"github.com/docker/docker-language-server/internal/pkg/cli/metadata"
)

var logLevel slog.LevelVar

type RootCmd struct {
	*cobra.Command
	debug   bool
	verbose bool
}

// Creates a new RootCmd
// params:
//
//	commandName: what to call the base command in examples (e.g., "docker-language-server")
func NewRootCmd(commandName string) *RootCmd {
	cmd := RootCmd{
		Command: &cobra.Command{
			Use:     commandName,
			Short:   "Language server for Docker",
			Version: metadata.Version,
		},
	}

	cmd.PersistentFlags().BoolVar(&cmd.debug, "debug", false, "Enable debug logging")
	cmd.PersistentFlags().BoolVar(&cmd.verbose, "verbose", false, "Enable verbose logging")

	cmd.PersistentPreRun = func(cc *cobra.Command, args []string) {
		if cmd.debug {
			logLevel.Set(slog.LevelDebug)
		} else if cmd.verbose {
			logLevel.Set(slog.LevelInfo)
		}
	}

	cmd.AddCommand(newStartCmd(commandName).Command)

	return &cmd
}

func Execute() {
	logLevel.Set(slog.LevelError)

	logger := slog.New(slog.NewJSONHandler(
		os.Stderr,
		&slog.HandlerOptions{
			AddSource: true,
			Level:     &logLevel,
		},
	))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	setupSignalHandler(cancel)

	err := NewRootCmd("docker-language-server").ExecuteContext(ctx)
	if err != nil {
		if !isCobraError(err) {
			logger.Error("fatal error", "error", err)
		}
		os.Exit(1)
	}
}

func setupSignalHandler(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			if sig == os.Interrupt {
				// TODO(milas): give open conns a grace period to close gracefully
				cancel()
				os.Exit(0)
			}
		}
	}()
}

func isCobraError(err error) bool {
	// Cobra doesn't give us a good way to distinguish between Cobra errors
	// (e.g. invalid command/args) and app errors, so ignore them manually
	// to avoid logging out scary stack traces for benign invocation issues
	return strings.Contains(err.Error(), "unknown flag")
}
