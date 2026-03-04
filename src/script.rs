use std::fs;
use std::io;
use std::path::Path;

const JORIN_SHEBANG: &str = "jorin";

#[derive(Debug, Clone, Copy, Eq, PartialEq)]
pub enum PromptMode {
    Auto,
    Text,
    File,
}

pub fn resolve_prompt(
    args: &[String],
    mode: PromptMode,
) -> Result<(String, Vec<String>), io::Error> {
    if args.is_empty() {
        if mode == PromptMode::File {
            return Err(io::Error::new(
                io::ErrorKind::InvalidInput,
                "prompt file required",
            ));
        }
        return Ok((String::new(), vec![]));
    }

    match mode {
        PromptMode::Text => Ok((args.join(" "), vec![])),
        PromptMode::File => {
            let prompt = load_prompt_file(&args[0], true)?
                .map(|(prompt, _)| prompt)
                .ok_or_else(|| {
                    io::Error::new(io::ErrorKind::InvalidInput, "prompt file required")
                })?;
            Ok((prompt, args[1..].to_vec()))
        }
        PromptMode::Auto => {
            if let Some((prompt, ok)) = load_prompt_file(&args[0], false)? {
                if ok {
                    return Ok((prompt, args[1..].to_vec()));
                }
            }
            Ok((args.join(" "), vec![]))
        }
    }
}

pub fn load_prompt_file(path: &str, require: bool) -> Result<Option<(String, bool)>, io::Error> {
    let p = Path::new(path);

    if !p.exists() {
        if !require {
            return Ok(None);
        }
        return Err(io::Error::new(
            io::ErrorKind::NotFound,
            format!("No such file or directory: {path}"),
        ));
    }

    if p.is_dir() {
        if require {
            return Err(io::Error::new(
                io::ErrorKind::InvalidInput,
                format!("prompt file is a directory: {path}"),
            ));
        }
        return Ok(None);
    }

    let data = fs::read_to_string(p)?;
    Ok(Some((parse_prompt_file(&data), true)))
}

pub fn parse_prompt_file(content: &str) -> String {
    if let Some((header, body)) = content.split_once('\n') {
        let header = header.trim_end_matches('\r');
        if header.starts_with("#!") && header.contains(JORIN_SHEBANG) {
            return body.trim_start_matches(['\r', '\n']).to_string();
        }
    } else {
        let header = content.trim_end_matches('\r');
        if header.starts_with("#!") && header.contains(JORIN_SHEBANG) {
            return String::new();
        }
    }

    content.to_string()
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs;

    #[test]
    fn test_load_script_prompt() {
        let temp = tempfile::tempdir().expect("tempdir");
        let path = temp.path().join("review-code.jorin");
        fs::write(
            &path,
            "#!/usr/bin/env jorin\nEnsure SOLID principles are followed.\n",
        )
        .expect("write");

        let loaded = load_prompt_file(path.to_str().expect("utf8 path"), true).expect("load");
        let (prompt, ok) = loaded.expect("expected file");
        assert!(ok);
        assert_eq!(prompt, "Ensure SOLID principles are followed.\n");
    }

    #[test]
    fn test_load_prompt_file_plain_text() {
        let temp = tempfile::tempdir().expect("tempdir");
        let path = temp.path().join("plain.txt");
        fs::write(&path, "just text\n").expect("write");

        let loaded = load_prompt_file(path.to_str().expect("utf8 path"), true).expect("load");
        let (prompt, ok) = loaded.expect("expected file");
        assert!(ok);
        assert_eq!(prompt, "just text\n");
    }

    #[test]
    fn test_resolve_prompt_auto_uses_prompt_file() {
        let temp = tempfile::tempdir().expect("tempdir");
        let path = temp.path().join("script.jorin");
        fs::write(&path, "#!/usr/bin/env jorin\nRun checks.").expect("write");

        let args = vec![
            path.to_str().expect("utf8 path").to_string(),
            "alpha".to_string(),
            "beta".to_string(),
        ];
        let (prompt, remaining) = resolve_prompt(&args, PromptMode::Auto).expect("resolve");
        assert_eq!(prompt, "Run checks.");
        assert_eq!(remaining, vec!["alpha", "beta"]);
    }

    #[test]
    fn test_resolve_prompt_forced_text() {
        let temp = tempfile::tempdir().expect("tempdir");
        let path = temp.path().join("prompt.txt");
        fs::write(&path, "Use the file.").expect("write");

        let args = vec![
            path.to_str().expect("utf8 path").to_string(),
            "extra".to_string(),
        ];
        let (prompt, remaining) = resolve_prompt(&args, PromptMode::Text).expect("resolve");
        assert_eq!(prompt, format!("{} extra", path.to_string_lossy()));
        assert!(remaining.is_empty());
    }

    #[test]
    fn test_resolve_prompt_requires_file() {
        let args = vec!["missing.txt".to_string()];
        assert!(resolve_prompt(&args, PromptMode::File).is_err());
    }
}
