use std::io::Read;

use clap::Parser;

use jorin::cli::Cli;
use jorin::openai::chat_session;
use jorin::prompt::system_prompt;
use jorin::script::resolve_prompt;
use jorin::types::{Message, Policy};

fn main() {
    let cli = Cli::parse();

    if cli.version {
        println!("{}", env!("CARGO_PKG_VERSION"));
        return;
    }

    let prompt_mode = match cli.prompt_mode() {
        Ok(mode) => mode,
        Err(err) => {
            eprintln!("ERR: {err}");
            std::process::exit(2);
        }
    };

    let (prompt, script_args) = match resolve_prompt(&cli.args, prompt_mode) {
        Ok(value) => value,
        Err(err) => {
            eprintln!("ERR: {err}");
            std::process::exit(1);
        }
    };

    let stdin_text = read_stdin_if_piped();
    let full_prompt = build_prompt(&prompt, &script_args, &stdin_text);

    if cli.repl || full_prompt.trim().is_empty() {
        eprintln!(
            "REPL mode is not implemented yet in Rust; pass a prompt to run single-shot mode."
        );
        std::process::exit(2);
    }

    let policy = Policy {
        readonly: cli.readonly,
        dry_shell: cli.dry_shell,
        allow: cli.allow,
        deny: cli.deny,
        cwd: cli.cwd,
    };

    let messages = vec![
        Message {
            role: "system".to_string(),
            content: Some(system_prompt()),
            name: None,
            tool_call_id: None,
            tool_calls: vec![],
        },
        Message {
            role: "user".to_string(),
            content: Some(full_prompt),
            name: None,
            tool_call_id: None,
            tool_calls: vec![],
        },
    ];

    match chat_session(&cli.model, messages, &policy) {
        Ok(output) => println!("{output}"),
        Err(err) => {
            eprintln!("ERR: {err}");
            std::process::exit(1);
        }
    }
}

fn read_stdin_if_piped() -> String {
    let mut buf = String::new();
    let mut stdin = std::io::stdin();
    if stdin.read_to_string(&mut buf).is_ok() {
        return buf;
    }
    String::new()
}

fn build_prompt(prompt: &str, args: &[String], stdin: &str) -> String {
    let trimmed_prompt = prompt.trim();
    let mut parts: Vec<String> = Vec::new();

    if !trimmed_prompt.is_empty() {
        parts.push(prompt.to_string());
    }
    if !args.is_empty() {
        parts.push(format!("Arguments: {}", args.join(" ")));
    }

    let stdin_trimmed = stdin.trim_end_matches('\n');
    if trimmed_prompt.is_empty() && args.is_empty() && !stdin_trimmed.is_empty() {
        return stdin_trimmed.to_string();
    }
    if !stdin_trimmed.is_empty() {
        parts.push(format!("Stdin:\n{stdin_trimmed}"));
    }

    parts.join("\n\n")
}
