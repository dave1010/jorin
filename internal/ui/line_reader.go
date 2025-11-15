package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/peterh/liner"
)

// LineReader abstracts reading single lines with proper terminal editing.
// Implementations must behave like bufio.Scanner for non-interactive inputs
// (e.g. when in/out are not terminals) so StartREPL remains testable.
type LineReader interface {
	ReadLine(prompt string) (string, error)
	Close() error
	// If supported, allow adding history lines
	AppendHistory(lines []string)
}

// NewLineReader returns a LineReader appropriate for the provided io.Reader/io.Writer.
// If input/output appear to be terminals it returns a real line editor using
// peterh/liner (supports left/right arrow, history navigation, etc.). Otherwise
// it falls back to a simple scanner-based implementation that reads from the
// provided reader.
func NewLineReader(in io.Reader, out io.Writer) LineReader {
	// detect terminal on both sides; require both so piping still works
	if fi, ok := in.(*os.File); ok {
		if fo, ok2 := out.(*os.File); ok2 {
			if isatty.IsTerminal(fi.Fd()) && isatty.IsTerminal(fo.Fd()) {
				l := liner.NewLiner()
				l.SetCtrlCAborts(true)
				l.SetMultiLineMode(false)
				// configure basic tab completion placeholder (none)
				return &linerReader{l: l}
			}
		}
	}
	// fallback scanner
	return &scannerReader{scanner: bufio.NewScanner(in), out: out}
}

// scannerReader implements LineReader using bufio.Scanner for non-tty input.
type scannerReader struct {
	scanner *bufio.Scanner
	out     io.Writer
}

func (s *scannerReader) ReadLine(prompt string) (string, error) {
	// mirror previous behavior: write prompt then read a line
	if _, err := fmt.Fprint(s.out, prompt); err != nil {
		return "", err
	}
	if s.scanner.Scan() {
		return s.scanner.Text(), nil
	}
	if err := s.scanner.Err(); err != nil {
		return "", err
	}
	return "", io.EOF
}

func (s *scannerReader) Close() error                 { return nil }
func (s *scannerReader) AppendHistory(lines []string) {}

// linerReader wraps peterh/liner
type linerReader struct {
	l *liner.State
}

func (lr *linerReader) ReadLine(prompt string) (string, error) {
	// liner expects prompt to be a string; ensure no trailing newline
	p := strings.TrimRight(prompt, "\n")
	line, err := lr.l.Prompt(p)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return "", io.EOF
		}
		return "", err
	}
	// add non-empty lines into liner's history so arrow up works in-session
	if strings.TrimSpace(line) != "" {
		lr.l.AppendHistory(line)
	}
	return line, nil
}

func (lr *linerReader) Close() error {
	return lr.l.Close()
}

func (lr *linerReader) AppendHistory(lines []string) {
	// append older entries first so they appear in chronological order
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			lr.l.AppendHistory(l)
		}
	}
}
