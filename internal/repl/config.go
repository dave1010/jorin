package repl

// Config holds REPL configuration.
type Config struct {
	Prompt              string // prompt string printed before input
	CommandPrefix       string // prefix for slash commands, default '/'
	EscapePrefix        string // prefix to escape command prefix, default '\\'
	MultilineTerminator string // not used yet
}

func DefaultConfig() *Config {
	return &Config{Prompt: "> ", CommandPrefix: "/", EscapePrefix: "\\"}
}
