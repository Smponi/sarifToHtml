package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	htmlreport "sarif-html/internal/html"
	"sarif-html/internal/report"
	"sarif-html/internal/sarif"
)

const version = "0.1.0"

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
	flags := flag.NewFlagSet("sarif-html", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	outPath := flags.String("out", "report.html", "output HTML file path, or - for stdout")
	title := flags.String("title", "SARIF HTML Report", "report title")
	repoURL := flags.String("repo-url", "", "repository URL for source links")
	revision := flags.String("revision", "", "repository revision, branch, tag, or commit for source links")
	sourceURLTemplate := flags.String("source-url-template", "", "URL template for source links; supports {path}, {line}, {lineFragment}, and {revision}")
	sourceRoots := sourceRootFlag{values: []string{"."}}
	flags.Var(&sourceRoots, "source-root", "source root used to load missing snippets; may be repeated")
	failOn := flags.String("fail-on", "", "exit with code 2 when a finding at or above severity exists: error, warning, note, none")
	showVersion := flags.Bool("version", false, "print version")

	if err := flags.Parse(reorderFlags(args)); err != nil {
		return err
	}

	if *showVersion {
		fmt.Println(version)
		return nil
	}

	inputs := flags.Args()
	if len(inputs) == 0 {
		return fmt.Errorf("at least one SARIF input file is required")
	}

	reports := make([]report.Report, 0, len(inputs))
	for _, input := range inputs {
		file, err := os.Open(input)
		if err != nil {
			return fmt.Errorf("open %s: %w", input, err)
		}
		log, err := sarif.Parse(file)
		closeErr := file.Close()
		if err != nil {
			return fmt.Errorf("parse %s: %w", input, err)
		}
		if closeErr != nil {
			return fmt.Errorf("close %s: %w", input, closeErr)
		}
		reports = append(reports, report.FromSARIF(log, filepath.Base(input)))
	}

	reportData := report.Merge(*title, reports...)
	if err := report.HydrateSnippets(&reportData, sourceRoots.values); err != nil {
		return err
	}
	linkOptions := sourceLinkOptions(*repoURL, *revision, *sourceURLTemplate)
	linkOptions.Title = *title
	output, err := htmlreport.Render(reportData, linkOptions)
	if err != nil {
		return err
	}

	if *outPath == "-" {
		if _, err := os.Stdout.Write(output); err != nil {
			return fmt.Errorf("write stdout: %w", err)
		}
	} else {
		if err := os.WriteFile(*outPath, output, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", *outPath, err)
		}
	}

	if strings.TrimSpace(*failOn) != "" {
		for _, finding := range reportData.Findings {
			if report.MeetsThreshold(finding.Level, *failOn) {
				return exitError{code: 2, message: fmt.Sprintf("finding %s meets fail-on threshold %s", finding.ID, *failOn)}
			}
		}
	}

	return nil
}

func reorderFlags(args []string) []string {
	valueFlags := map[string]bool{
		"out":                 true,
		"title":               true,
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

func (flag *sourceRootFlag) String() string {
	return strings.Join(flag.values, ",")
}

func (flag *sourceRootFlag) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	flag.values = append(flag.values, value)
	return nil
}
