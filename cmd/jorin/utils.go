package main

import "github.com/dave1010/jorin/internal/tools"

func dirOrDot(p string) string { return tools.DirOrDot(p) }

func tail(s string, n int) string { return tools.Tail(s, n) }

func preview(s string, n int) string { return tools.Preview(s, n) }
