package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/steffenrumpf/hdc/internal/checker"
	"github.com/steffenrumpf/hdc/internal/config"
	"github.com/steffenrumpf/hdc/internal/parser"
	"github.com/steffenrumpf/hdc/internal/report"
	"github.com/steffenrumpf/hdc/internal/repository"
)

// Injected at build time via ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var cfgFile string

	root := &cobra.Command{
		Use:   "hdc",
		Short: "Helmfile Dependency Checker",
		Long:  "hdc verifies that Helm chart dependencies in helmfiles are up-to-date and actively maintained.",
	}

	root.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: .helmfile-checker.yaml)")
	root.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")
	root.PersistentFlags().String("log-format", "text", "log format (text, json)")

	if err := viper.BindPFlag("log.level", root.PersistentFlags().Lookup("log-level")); err != nil {
		slog.Error("failed to bind log-level flag", "error", err)
	}

	if err := viper.BindPFlag("log.format", root.PersistentFlags().Lookup("log-format")); err != nil {
		slog.Error("failed to bind log-format flag", "error", err)
	}

	root.AddCommand(newCheckCmd(cfgFile))
	root.AddCommand(newVersionCmd())

	return root
}

func newCheckCmd(cfgFile string) *cobra.Command {
	var (
		outputFile    string
		ignoreSkipped bool
		maxAge        int
		concurrent    int
		timeout       int
	)

	cmd := &cobra.Command{
		Use:   "check <helmfile-path>",
		Short: "Check helmfile dependencies for outdated or unmaintained charts",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.InitConfig(cfgFile)
			if err != nil {
				return err
			}

			config.InitLogger(cfg)

			if err := bindCheckFlags(cmd, cfg); err != nil {
				return err
			}

			return runCheck(args[0], cfg)
		},
	}

	cmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config file (default: .helmfile-checker.yaml)")
	cmd.Flags().StringP("output", "o", "json", "output format (json, markdown, html)")
	cmd.Flags().StringVarP(&outputFile, "output-file", "f", "", "write report to file instead of stdout")
	cmd.Flags().BoolVarP(&ignoreSkipped, "ignore-skipped", "i", false, "omit skipped releases from report output")
	cmd.Flags().IntVarP(&maxAge, "max-age", "m", 12, "maximum chart age in months before flagged as unmaintained")
	cmd.Flags().IntVarP(&concurrent, "concurrent", "n", 5, "number of concurrent repository queries")
	cmd.Flags().IntVarP(&timeout, "timeout", "t", 30, "repository request timeout in seconds")

	return cmd
}

func bindCheckFlags(cmd *cobra.Command, cfg *config.Config) error {
	flags := []struct {
		flag string
		key  string
	}{
		{"output", "output.format"},
		{"output-file", "output.file"},
		{"ignore-skipped", "output.ignore_skipped"},
		{"max-age", "checker.max_age_months"},
		{"concurrent", "checker.concurrent_requests"},
		{"timeout", "repositories.timeout_seconds"},
	}

	for _, f := range flags {
		if err := viper.BindPFlag(f.key, cmd.Flags().Lookup(f.flag)); err != nil {
			return fmt.Errorf("failed to bind flag %s: %w", f.flag, err)
		}
	}

	cfg.Output.Format = viper.GetString("output.format")
	cfg.Output.File = viper.GetString("output.file")
	cfg.Output.IgnoreSkipped = viper.GetBool("output.ignore_skipped")
	cfg.Checker.MaxAgeMonths = viper.GetInt("checker.max_age_months")
	cfg.Checker.ConcurrentRequests = viper.GetInt("checker.concurrent_requests")
	cfg.Repositories.TimeoutSeconds = viper.GetInt("repositories.timeout_seconds")

	return nil
}

func runCheck(helmfilePath string, cfg *config.Config) error {
	slog.Info("parsing helmfile", "path", helmfilePath)

	hf, err := parser.New().Parse(helmfilePath)
	if err != nil {
		return err
	}

	slog.Info("found releases", "count", len(hf.Releases))

	timeout := time.Duration(cfg.Repositories.TimeoutSeconds) * time.Second
	repoClient := repository.NewDefault(timeout)

	chk := checker.New(repoClient, checker.Config{
		MaxAgeMonths:       cfg.Checker.MaxAgeMonths,
		ConcurrentRequests: cfg.Checker.ConcurrentRequests,
		ExcludeCharts:      cfg.Exclude.Charts,
		ExcludeRepos:       cfg.Exclude.Repositories,
	})

	result, err := chk.Check(hf)
	if err != nil {
		return err
	}

	writer, err := report.New(cfg.Output.Format, cfg.Output.IgnoreSkipped)
	if err != nil {
		return err
	}

	out := os.Stdout
	var f *os.File
	if cfg.Output.File != "" {
		var err error
		f, err = os.Create(cfg.Output.File)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		out = f
	}

	if err := writer.Write(out, result); err != nil {
		return err
	}

	// Close file if it was opened
	if f != nil {
		if err := f.Close(); err != nil {
			return fmt.Errorf("failed to close output file: %w", err)
		}
	}

	// Exit with severity-based exit code
	exitCode := result.ExitCode()
	if exitCode > 0 {
		slog.Warn("issues found in helmfile dependencies", "exit_code", exitCode)
	}
	os.Exit(exitCode)
	return nil // This line is never reached but satisfies the compiler
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print hdc version information",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("hdc %s (commit: %s, built: %s)\n", version, commit, date)
		},
	}
}
