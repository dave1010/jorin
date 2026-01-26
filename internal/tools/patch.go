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

type hunk struct {
	oldStart int
	oldLen   int
	newStart int
	newLen   int
	lines    []string
}

type parsedPatch struct {
	op       operation
	filePath string
	hunks    []hunk
}

func parsePatch(patch string) (*parsedPatch, error) {
	scanner := bufio.NewScanner(strings.NewReader(patch))
	var line1, line2 string

	// Skip until we find "--- "
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "--- ") {
			line1 = line
			break
		}
	}

	if line1 == "" {
		return nil, errors.New("invalid patch format: missing '--- ' header")
	}

	if scanner.Scan() {
		line2 = scanner.Text()
	} else {
		return nil, errors.New("invalid patch format: missing second line after '--- '")
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
			return nil, fmt.Errorf("file paths in patch header do not match: %q (from %q) vs %q (from %q). Ensure both headers use the same file path", p.filePath, line1, fp2, line2)
		}
	}

	var currentHunk *hunk
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "@@") {
			if currentHunk != nil {
				p.hunks = append(p.hunks, *currentHunk)
			}
			currentHunk = &hunk{}
			_, err := fmt.Sscanf(line, "@@ -%d,%d +%d,%d @@", &currentHunk.oldStart, &currentHunk.oldLen, &currentHunk.newStart, &currentHunk.newLen)
			if err != nil {
				// try alternative format @@ -%d +%d @@
				_, err = fmt.Sscanf(line, "@@ -%d +%d @@", &currentHunk.oldStart, &currentHunk.newStart)
				if err != nil {
					return nil, fmt.Errorf("malformed hunk header: %q. Expected format '@@ -oldStart,oldLen +newStart,newLen @@'", line)
				}
				currentHunk.oldLen = 1
				currentHunk.newLen = 1
			}
			continue
		}
		if currentHunk != nil {
			currentHunk.lines = append(currentHunk.lines, line)
		}
	}
	if currentHunk != nil {
		p.hunks = append(p.hunks, *currentHunk)
	}

	return &p, nil
}

func parseFilePath(line string) (string, error) {
	parts := strings.SplitN(line, " ", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid patch header line: %q", line)
	}
	pathPart := strings.TrimSpace(parts[1])
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
		return createFile(p.filePath, p.hunks)
	case opUpdate:
		return updateFile(p.filePath, p.hunks)
	case opDelete:
		return deleteFile(p.filePath)
	}
	return errors.New("unknown operation")
}

func createFile(path string, hunks []hunk) error {
	var content strings.Builder
	for _, h := range hunks {
		for _, line := range h.lines {
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
	}

	return os.WriteFile(path, []byte(content.String()), 0644)
}

func updateFile(path string, hunks []hunk) error {
	f, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	originalLines := strings.Split(string(f), "\n")
	if len(originalLines) > 0 && originalLines[len(originalLines)-1] == "" {
		originalLines = originalLines[:len(originalLines)-1]
	}

	currentLines := originalLines

	// Track the cumulative shift in line numbers caused by previous hunks
	globalOffset := 0

	for i, h := range hunks {
		// Calculate where we expect the hunk to apply based on its header and previous shifts
		expectedStart := h.oldStart - 1 + globalOffset
		if expectedStart < 0 {
			expectedStart = 0
		}

		// Find the actual application point
		applyAt, err := findHunkLocation(currentLines, h, expectedStart)
		if err != nil {
			return fmt.Errorf("failed to apply hunk %d: %w", i+1, err)
		}

		// Apply the hunk at the found location
		var nextLines []string
		nextLines = append(nextLines, currentLines[:applyAt]...)

		for _, line := range h.lines {
			if strings.HasPrefix(line, "+") {
				nextLines = append(nextLines, line[1:])
			} else if strings.HasPrefix(line, " ") {
				// Keep existing line
				if applyAt < len(currentLines) {
					nextLines = append(nextLines, currentLines[applyAt])
					applyAt++
				}
			} else if strings.HasPrefix(line, "-") {
				// Skip existing line (delete)
				applyAt++
			}
		}

		// Append the rest of the file
		if applyAt < len(currentLines) {
			nextLines = append(nextLines, currentLines[applyAt:]...)
		}

		// Update global offset for next hunk
		// The shift is newLen - oldLen
		// However, since we are re-searching for the location for every hunk,
		// strict globalOffset tracking is less critical for finding the location,
		// but `findHunkLocation` uses expectedStart as a hint.
		// Let's rely on the difference between the found location and where we ended up.

		hunkOldLen := 0
		hunkNewLen := 0
		for _, line := range h.lines {
			if strings.HasPrefix(line, " ") {
				hunkOldLen++
				hunkNewLen++
			} else if strings.HasPrefix(line, "-") {
				hunkOldLen++
			} else if strings.HasPrefix(line, "+") {
				hunkNewLen++
			}
		}

		// Adjust globalOffset.
		// Actually, since we rebuild currentLines, the indices for the *next* hunk
		// in the *original* file (h.oldStart) need to be adjusted by how much
		// we've changed the file size so far.
		globalOffset += (hunkNewLen - hunkOldLen)

		currentLines = nextLines
	}

	return os.WriteFile(path, []byte(strings.Join(currentLines, "\n")+"\n"), 0644)
}

func findHunkLocation(lines []string, h hunk, expectedStart int) (int, error) {
	// Construct the block of lines we are looking for (context + deleted lines)
	var searchBlock []string
	for _, line := range h.lines {
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "-") {
			searchBlock = append(searchBlock, line[1:])
		}
	}

	if len(searchBlock) == 0 {
		// Pure addition? It applies at expectedStart
		return expectedStart, nil
	}

	// 1. Try exact match at expectedStart
	if matchAt(lines, searchBlock, expectedStart) {
		return expectedStart, nil
	}

	// 2. Search locally around expectedStart? Or just global search?
	// Let's do a global search and find the closest match to expectedStart.
	var candidates []int
	for i := 0; i <= len(lines)-len(searchBlock); i++ {
		if matchAt(lines, searchBlock, i) {
			candidates = append(candidates, i)
		}
	}

	if len(candidates) == 0 {
		// Construct a detailed error message showing what was expected vs found at expectedStart
		// Limit the output to first few lines to avoid huge error messages
		details := ""
		if expectedStart < len(lines) {
			limit := 3
			if len(searchBlock) < limit {
				limit = len(searchBlock)
			}
			details = fmt.Sprintf("\nExpected first %d lines:\n", limit)
			for i := 0; i < limit; i++ {
				details += fmt.Sprintf("  %q\n", searchBlock[i])
			}
			details += fmt.Sprintf("Found at line %d:\n", expectedStart+1)
			for i := 0; i < limit && expectedStart+i < len(lines); i++ {
				details += fmt.Sprintf("  %q\n", lines[expectedStart+i])
			}
		} else {
			details = fmt.Sprintf("\nExpected start at line %d, but file has only %d lines.", expectedStart+1, len(lines))
		}
		return -1, fmt.Errorf("hunk context not found in file%s", details)
	}

	// Find closest candidate
	bestCandidate := candidates[0]
	minDist := abs(candidates[0] - expectedStart)

	for _, c := range candidates {
		dist := abs(c - expectedStart)
		if dist < minDist {
			minDist = dist
			bestCandidate = c
		}
	}

	return bestCandidate, nil
}

func matchAt(lines []string, block []string, start int) bool {
	if start < 0 || start+len(block) > len(lines) {
		return false
	}
	for i, line := range block {
		if lines[start+i] != line {
			return false
		}
	}
	return true
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func deleteFile(path string) error {
	return os.Remove(path)
}
