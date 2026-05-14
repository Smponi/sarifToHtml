package report

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const maxSnippetLines = 20

// HydrateSnippets loads source snippets for findings whose SARIF result did not
// already include one. It only reads relative paths below the provided roots.
func HydrateSnippets(reportData *Report, roots []string) error {
	if reportData == nil {
		return nil
	}
	roots = cleanSourceRoots(roots)
	if len(roots) == 0 {
		return nil
	}

	for index := range reportData.Findings {
		finding := &reportData.Findings[index]
		if strings.TrimSpace(finding.Snippet) != "" || finding.Path == "" || finding.StartLine <= 0 {
			continue
		}
		snippet, err := readSnippet(roots, finding.Path, finding.StartLine, finding.EndLine)
		if err != nil {
			return err
		}
		finding.Snippet = snippet
	}
	return nil
}

func cleanSourceRoots(roots []string) []string {
	cleaned := make([]string, 0, len(roots))
	seen := map[string]bool{}
	for _, root := range roots {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		absolute, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		absolute = filepath.Clean(absolute)
		if !seen[absolute] {
			cleaned = append(cleaned, absolute)
			seen[absolute] = true
		}
	}
	return cleaned
}

func readSnippet(roots []string, findingPath string, startLine, endLine int) (string, error) {
	cleanPath := strings.TrimLeft(filepath.Clean(strings.ReplaceAll(findingPath, "\\", string(filepath.Separator))), string(filepath.Separator))
	if cleanPath == "." || cleanPath == "" || isAbsolutePath(strings.ReplaceAll(findingPath, "\\", "/")) {
		return "", nil
	}

	for _, root := range roots {
		candidate := filepath.Join(root, cleanPath)
		if !isBelowRoot(root, candidate) {
			continue
		}
		snippet, found, err := readSnippetFromFile(candidate, startLine, endLine)
		if err != nil {
			return "", err
		}
		if found {
			return snippet, nil
		}
	}
	return "", nil
}

func isBelowRoot(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	return relative == "." || (!strings.HasPrefix(relative, ".."+string(filepath.Separator)) && relative != "..")
}

func readSnippetFromFile(path string, startLine, endLine int) (string, bool, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("read snippet from %s: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	if endLine < startLine {
		endLine = startLine
	}
	if endLine-startLine+1 > maxSnippetLines {
		endLine = startLine + maxSnippetLines - 1
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		if lineNumber < startLine {
			continue
		}
		if lineNumber > endLine {
			break
		}
		lines = append(lines, strings.TrimRight(scanner.Text(), "\r"))
	}
	if err := scanner.Err(); err != nil {
		return "", false, fmt.Errorf("read snippet from %s: %w", path, err)
	}
	if len(lines) == 0 {
		return "", false, nil
	}
	return strings.Join(lines, "\n"), true, nil
}
