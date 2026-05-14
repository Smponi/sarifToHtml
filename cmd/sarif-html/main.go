package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	htmlreport "sarif-html/internal/html"
	"sarif-html/internal/report"
	"sarif-html/internal/sarif"
)

const version = "0.1.0"

// cliConfig is the fully parsed command configuration. Keeping flag parsing
// separate from execution makes the command easier to test and keeps run focused
// on the report pipeline.
type cliConfig struct {
	outPath           string
	title             string
	templatePath      string
	repoURL           string
	revision          string
	sourceURLTemplate string
	sourceRoots       []string
	failOn            string
	showVersion       bool
	inputs            []string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "sarif-html:", err)
		var exitErr exitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.code)
		}
		os.Exit(1)
	}
}

func run(args []string) error {
	config, err := parseConfig(args)
	if err != nil {
		return err
	}
	if config.showVersion {
		return printVersion(os.Stdout)
	}
	if len(config.inputs) == 0 {
		return fmt.Errorf("at least one SARIF input file is required")
	}

	reports, err := loadReports(config.inputs)
	if err != nil {
		return err
	}

	reportData := report.Merge(config.title, reports...)
	if err := report.HydrateSnippets(&reportData, config.sourceRoots); err != nil {
		return err
	}

	renderOptions := sourceLinkOptions(config.repoURL, config.revision, config.sourceURLTemplate)
	renderOptions.Title = config.title
	renderOptions.TemplatePath = config.templatePath

	output, err := htmlreport.Render(reportData, renderOptions)
	if err != nil {
		return err
	}
	if err := writeOutput(config.outPath, output); err != nil {
		return err
	}

	return enforceFailOn(config.failOn, reportData.Findings)
}

// parseConfig owns all CLI flag defaults and compatibility behavior. The rest
// of the command should consume cliConfig instead of reaching back into flag
// pointers, which keeps command execution deterministic in tests.
func parseConfig(args []string) (cliConfig, error) {
	flags := flag.NewFlagSet("sarif-html", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	outPath := flags.String("out", "report.html", "output HTML file path, or - for stdout")
	title := flags.String("title", "SARIF HTML Report", "report title")
	templatePath := flags.String("template", "", "custom Go html/template file or directory; default uses the built-in report template")
	repoURL := flags.String("repo-url", "", "repository URL for source links")
	revision := flags.String("revision", "", "repository revision, branch, tag, or commit for source links")
	sourceURLTemplate := flags.String("source-url-template", "", "URL template for source links; supports {path}, {line}, {lineFragment}, and {revision}")
	sourceRoots := sourceRootFlag{values: []string{"."}}
	flags.Var(&sourceRoots, "source-root", "source root used to load missing snippets; may be repeated")
	failOn := flags.String("fail-on", "", "exit with code 2 when a finding at or above severity exists: error, warning, note, none")
	showVersion := flags.Bool("version", false, "print version")

	if err := flags.Parse(reorderFlags(args)); err != nil {
		return cliConfig{}, err
	}

	return cliConfig{
		outPath:           *outPath,
		title:             *title,
		templatePath:      *templatePath,
		repoURL:           *repoURL,
		revision:          *revision,
		sourceURLTemplate: *sourceURLTemplate,
		sourceRoots:       sourceRoots.values,
		failOn:            *failOn,
		showVersion:       *showVersion,
		inputs:            flags.Args(),
	}, nil
}

func printVersion(w io.Writer) error {
	if _, err := fmt.Fprintln(w, version); err != nil {
		return fmt.Errorf("write version: %w", err)
	}
	return nil
}

// loadReports preserves the input file name as the report source label while
// letting each file fail with path-specific context.
func loadReports(inputs []string) ([]report.Report, error) {
	reports := make([]report.Report, 0, len(inputs))
	for _, input := range inputs {
		sarifReport, err := loadReport(input)
		if err != nil {
			return nil, err
		}
		reports = append(reports, sarifReport)
	}
	return reports, nil
}

// loadReport closes the input explicitly after parsing so close failures can be
// reported to users writing from networked or virtual filesystems.
func loadReport(input string) (report.Report, error) {
	file, err := os.Open(input)
	if err != nil {
		return report.Report{}, fmt.Errorf("open %s: %w", input, err)
	}

	log, parseErr := sarif.Parse(file)
	closeErr := file.Close()
	if parseErr != nil {
		return report.Report{}, fmt.Errorf("parse %s: %w", input, parseErr)
	}
	if closeErr != nil {
		return report.Report{}, fmt.Errorf("close %s: %w", input, closeErr)
	}
	return report.FromSARIF(log, filepath.Base(input)), nil
}

// writeOutput centralizes the file/stdout split used by both normal report
// generation and tests.
func writeOutput(outPath string, output []byte) error {
	if outPath == "-" {
		if _, err := os.Stdout.Write(output); err != nil {
			return fmt.Errorf("write stdout: %w", err)
		}
		return nil
	}

	if err := os.WriteFile(outPath, output, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}
	return nil
}

// enforceFailOn runs after the report is written. CI users still get the HTML
// artifact even when the command exits with the threshold failure code.
func enforceFailOn(threshold string, findings []report.Finding) error {
	if strings.TrimSpace(threshold) == "" {
		return nil
	}
	for _, finding := range findings {
		if report.MeetsThreshold(finding.Level, threshold) {
			return exitError{code: 2, message: fmt.Sprintf("finding %s meets fail-on threshold %s", finding.ID, threshold)}
		}
	}
	return nil
}

// reorderFlags keeps the CLI friendly for artifact-style commands where users
// often write `sarif-html input.sarif --out report.html`. The standard flag
// package stops parsing at the first positional argument, so known flags are
// moved in front before parsing.
func reorderFlags(args []string) []string {
	valueFlags := map[string]bool{
		"out":                 true,
		"title":               true,
		"template":            true,
		"repo-url":            true,
		"revision":            true,
		"source-url-template": true,
		"source-root":         true,
		"fail-on":             true,
	}

	var flagArgs []string
	var positional []string
	for index := 0; index < len(args); index++ {
		arg := args[index]
		if arg == "--" {
			positional = append(positional, args[index+1:]...)
			break
		}
		if strings.HasPrefix(arg, "-") && arg != "-" {
			name := strings.TrimLeft(arg, "-")
			if equal := strings.Index(name, "="); equal >= 0 {
				flagArgs = append(flagArgs, arg)
				continue
			}
			flagArgs = append(flagArgs, arg)
			if valueFlags[name] && index+1 < len(args) {
				index++
				flagArgs = append(flagArgs, args[index])
			}
			continue
		}
		positional = append(positional, arg)
	}

	return append(flagArgs, positional...)
}

// sourceLinkOptions respects explicit link settings first, then infers a source
// URL template from common GitHub Actions and GitLab CI environment variables.
func sourceLinkOptions(repoURL, revision, sourceURLTemplate string) htmlreport.Options {
	options := htmlreport.Options{
		RepoURL:           repoURL,
		Revision:          revision,
		SourceURLTemplate: sourceURLTemplate,
	}
	if options.SourceURLTemplate != "" {
		return options
	}
	if options.RepoURL != "" && options.Revision != "" {
		return options
	}

	if githubRepository := os.Getenv("GITHUB_REPOSITORY"); githubRepository != "" {
		serverURL := firstNonEmpty(os.Getenv("GITHUB_SERVER_URL"), "https://github.com")
		ref := firstNonEmpty(os.Getenv("GITHUB_SHA"), os.Getenv("GITHUB_HEAD_REF"), os.Getenv("GITHUB_REF_NAME"))
		if ref != "" {
			options.Revision = ref
			options.SourceURLTemplate = strings.TrimSuffix(serverURL, "/") + "/" + githubRepository + "/blob/{revision}/{path}{lineFragment}"
			return options
		}
	}

	if projectURL := os.Getenv("CI_PROJECT_URL"); projectURL != "" {
		ref := firstNonEmpty(os.Getenv("CI_MERGE_REQUEST_SOURCE_BRANCH_SHA"), os.Getenv("CI_COMMIT_SHA"), os.Getenv("CI_COMMIT_REF_NAME"))
		if ref != "" {
			options.Revision = ref
			options.SourceURLTemplate = strings.TrimSuffix(projectURL, "/") + "/-/blob/{revision}/{path}{lineFragment}"
			return options
		}
	}

	return options
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

type exitError struct {
	code    int
	message string
}

func (err exitError) Error() string {
	return err.message
}

type sourceRootFlag struct {
	values []string
}

// String implements flag.Value for repeated --source-root values.
func (flag *sourceRootFlag) String() string {
	return strings.Join(flag.values, ",")
}

// Set appends non-empty source roots and keeps the default root intact. That
// lets users add roots without accidentally disabling the current directory.
func (flag *sourceRootFlag) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	flag.values = append(flag.values, value)
	return nil
}
