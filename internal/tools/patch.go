package tools

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

type operation int

const (
	opCreate operation = iota
	opUpdate
	opDelete
)

type parsedPatch struct {
	op       operation
	filePath string
	diff     string
}

func parsePatch(patch string) (*parsedPatch, error) {
	scanner := bufio.NewScanner(strings.NewReader(patch))
	var line1, line2 string

	if scanner.Scan() {
		line1 = scanner.Text()
	} else {
		return nil, errors.New("invalid patch format: empty patch")
	}

	if scanner.Scan() {
		line2 = scanner.Text()
	} else {
		return nil, errors.New("invalid patch format: missing second line")
	}

	var p parsedPatch
	var err error

	if strings.HasPrefix(line1, "--- /dev/null") {
		p.op = opCreate
		p.filePath, err = parseFilePath(line2)
		if err != nil {
			return nil, err
		}
	} else if strings.HasPrefix(line2, "+++ /dev/null") {
		p.op = opDelete
		p.filePath, err = parseFilePath(line1)
		if err != nil {
			return nil, err
		}
	} else {
		p.op = opUpdate
		p.filePath, err = parseFilePath(line1)
		if err != nil {
			return nil, err
		}
		// Sanity check
		fp2, err := parseFilePath(line2)
		if err != nil {
			return nil, err
		}
		if p.filePath != fp2 {
			return nil, fmt.Errorf("file paths in patch header do not match: %q vs %q", p.filePath, fp2)
		}
	}

	var diffLines []string
	var hunkHeaderSeen bool
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "@@") {
			hunkHeaderSeen = true
			continue
		}
		if hunkHeaderSeen {
			diffLines = append(diffLines, line)
		}
	}
	p.diff = strings.Join(diffLines, "\n")

	return &p, nil
}

func parseFilePath(line string) (string, error) {
	parts := strings.SplitN(line, " ", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid patch header line: %q", line)
	}
	pathPart := parts[1]
	if strings.HasPrefix(pathPart, "a/") || strings.HasPrefix(pathPart, "b/") {
		return pathPart[2:], nil
	}
	return pathPart, nil
}

func ApplyPatch(patch string) error {
	p, err := parsePatch(patch)
	if err != nil {
		return err
	}

	switch p.op {
	case opCreate:
		return createFile(p.filePath, p.diff)
	case opUpdate:
		return updateFile(p.filePath, p.diff)
	case opDelete:
		return deleteFile(p.filePath)
	}
	return errors.New("unknown operation")
}

func createFile(path, diff string) error {
	var content strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(diff))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "+") {
			content.WriteString(line[1:])
			content.WriteString("\n")
		} else if strings.HasPrefix(line, " ") {
			content.WriteString(line[1:])
			content.WriteString("\n")
		} else if !strings.HasPrefix(line, "-") {
			return fmt.Errorf("invalid diff content for create: %s", line)
		}
	}

	return os.WriteFile(path, []byte(content.String()), 0644)
}

func updateFile(path, diff string) error {
	f, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	originalLines := strings.Split(string(f), "\n")
	if len(originalLines) > 0 && originalLines[len(originalLines)-1] == "" {
		originalLines = originalLines[:len(originalLines)-1]
	}
	var newLines []string
	originalLineIndex := 0

	scanner := bufio.NewScanner(strings.NewReader(diff))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "+") {
			newLines = append(newLines, line[1:])
		} else if strings.HasPrefix(line, "-") {
			originalLineIndex++
		} else if strings.HasPrefix(line, " ") {
			if originalLineIndex < len(originalLines) {
				newLines = append(newLines, originalLines[originalLineIndex])
			}
			originalLineIndex++
		} else {
			return fmt.Errorf("invalid diff content for update: %s", line)
		}
	}

	for originalLineIndex < len(originalLines) {
		newLines = append(newLines, originalLines[originalLineIndex])
		originalLineIndex++
	}

	return os.WriteFile(path, []byte(strings.Join(newLines, "\n")+"\n"), 0644)
}

func deleteFile(path string) error {
	return os.Remove(path)
}
