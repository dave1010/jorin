use clap::Parser;

use crate::script::PromptMode;

#[derive(Debug, Parser)]
#[command(name = "jorin", about = "Jorin coding agent (Rust migration)")]
pub struct Cli {
    #[arg(long, default_value = "gpt-5-mini")]
    pub model: String,

    #[arg(long)]
    pub repl: bool,

    #[arg(long)]
    pub prompt: bool,

    #[arg(long = "prompt-file")]
    pub prompt_file: bool,

    #[arg(long)]
    pub readonly: bool,

    #[arg(long = "dry-shell")]
    pub dry_shell: bool,

    #[arg(long)]
    pub allow: Vec<String>,

    #[arg(long)]
    pub deny: Vec<String>,

    #[arg(long)]
    pub cwd: Option<String>,

    #[arg(long)]
    pub ralph: bool,

    #[arg(long = "ralph-max-tries", default_value_t = 8)]
    pub ralph_max_tries: u32,

    #[arg(long = "use-responses-api")]
    pub use_responses_api: bool,

    #[arg(long)]
    pub version: bool,

    pub args: Vec<String>,
}

impl Cli {
    pub fn prompt_mode(&self) -> Result<PromptMode, String> {
        if self.prompt && self.prompt_file {
            return Err("flag --prompt and --prompt-file cannot be used together".to_string());
        }

        if self.prompt {
            return Ok(PromptMode::Text);
        }
        if self.prompt_file {
            return Ok(PromptMode::File);
        }
        Ok(PromptMode::Auto)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use clap::CommandFactory;

    #[test]
    fn clap_debug_assert() {
        Cli::command().debug_assert();
    }
}
