package main

import (
	"errors"
	"io/fs"
	"os"
	"strings"
)

const jorinShebang = "jorin"

func resolvePrompt(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "", nil, nil
	}
	prompt, ok, err := loadScriptPrompt(args[0])
	if err != nil {
		return "", nil, err
	}
	if ok {
		return prompt, args[1:], nil
	}
	return stringJoin(args, " "), nil, nil
}

func loadScriptPrompt(path string) (string, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", false, nil
		}
		return "", false, err
	}

	content := string(data)
	lines := strings.SplitN(content, "\n", 2)
	header := strings.TrimSuffix(lines[0], "\r")
	if !strings.HasPrefix(header, "#!") || !strings.Contains(header, jorinShebang) {
		return "", false, nil
	}

	body := ""
	if len(lines) == 2 {
		body = strings.TrimLeft(lines[1], "\r\n")
	}
	return body, true, nil
}

func isTTY(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}
