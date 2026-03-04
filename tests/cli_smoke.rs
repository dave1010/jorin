use std::process::Command;

fn bin() -> Command {
    Command::new(env!("CARGO_BIN_EXE_jorin"))
}

#[test]
fn supports_version_flag() {
    let output = bin().arg("--version").output().expect("binary runs");
    assert!(output.status.success());
    assert!(String::from_utf8_lossy(&output.stdout).contains("0.1.0"));
}

#[test]
fn rejects_conflicting_prompt_flags() {
    let output = bin()
        .args(["--prompt", "--prompt-file", "input.txt"])
        .output()
        .expect("binary runs");
    assert_eq!(output.status.code(), Some(2));
    assert!(String::from_utf8_lossy(&output.stderr)
        .contains("flag --prompt and --prompt-file cannot be used together"));
}

#[test]
fn rejects_missing_prompt_without_repl() {
    let output = bin().output().expect("binary runs");
    assert_eq!(output.status.code(), Some(2));
    assert!(String::from_utf8_lossy(&output.stderr)
        .contains("REPL mode is not implemented yet in Rust"));
}
