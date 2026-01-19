package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

const jorinShebang = "jorin"

type promptMode int

const (
	promptModeAuto promptMode = iota
	promptModeText
	promptModeFile
)

func resolvePrompt(args []string, mode promptMode) (string, []string, error) {
	if len(args) == 0 {
		if mode == promptModeFile {
			return "", nil, errors.New("prompt file required")
		}
		return "", nil, nil
	}
	if mode == promptModeText {
		return stringJoin(args, " "), nil, nil
	}
	if mode == promptModeFile {
		prompt, _, err := loadPromptFile(args[0], true)
		if err != nil {
			return "", nil, err
		}
		return prompt, args[1:], nil
	}

	prompt, ok, err := loadPromptFile(args[0], false)
	if err != nil {
		return "", nil, err
	}
	if ok {
		return prompt, args[1:], nil
	}
	return stringJoin(args, " "), nil, nil
}

func loadPromptFile(path string, require bool) (string, bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if !require && errors.Is(err, fs.ErrNotExist) {
			return "", false, nil
		}
		return "", false, err
	}
	if info.IsDir() {
		if require {
			return "", false, fmt.Errorf("prompt file is a directory: %s", path)
		}
		return "", false, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", false, err
	}
	return parsePromptFile(string(data)), true, nil
}

func parsePromptFile(content string) string {
	lines := strings.SplitN(content, "\n", 2)
	header := strings.TrimSuffix(lines[0], "\r")
	if !strings.HasPrefix(header, "#!") || !strings.Contains(header, jorinShebang) {
		return content
	}

	body := ""
	if len(lines) == 2 {
		body = strings.TrimLeft(lines[1], "\r\n")
	}
	return body
}

func isTTY(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}
