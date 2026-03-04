use std::collections::HashMap;
use std::fs;
use std::io::Read;
use std::process::Command;

use anyhow::{anyhow, Context, Result};
use serde_json::{json, Value};

use crate::text::{dir_or_dot, preview, tail};
use crate::types::{Policy, Tool, ToolCall, ToolFunction};

const MAX_READ_FILE_BYTES: usize = 200_000;
const MAX_TOOL_OUTPUT_BYTES: usize = 8_000;

pub fn tools_manifest() -> Vec<Tool> {
    vec![
        tool(
            "shell",
            "Execute a shell command; returns stdout/stderr/returncode.",
            json!({"type":"object","properties":{"cmd":{"type":"string"}},"required":["cmd"]}),
        ),
        tool(
            "read_file",
            "Read a UTF-8 text file and return contents.",
            json!({"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}),
        ),
        tool(
            "write_file",
            "Write UTF-8 text to a file (creates/overwrites).",
            json!({"type":"object","properties":{"path":{"type":"string"},"text":{"type":"string"}},"required":["path","text"]}),
        ),
        tool(
            "http_get",
            "Fetch URL and return body text.",
            json!({"type":"object","properties":{"url":{"type":"string"}},"required":["url"]}),
        ),
        tool(
            "apply_patch",
            "Apply a unified diff patch using git apply.",
            json!({"type":"object","properties":{"patch":{"type":"string"}},"required":["patch"]}),
        ),
    ]
}

fn tool(name: &str, description: &str, parameters: Value) -> Tool {
    Tool {
        kind: "function".into(),
        function: ToolFunction {
            name: name.into(),
            description: description.into(),
            parameters,
        },
    }
}

pub fn execute_tool_call(tc: &ToolCall, policy: &Policy) -> Value {
    let parsed =
        serde_json::from_str::<HashMap<String, Value>>(&tc.function.arguments).or_else(|_| {
            serde_json::from_str::<String>(&tc.function.arguments)
                .map(|s| HashMap::from([("_raw".to_string(), Value::String(s))]))
        });

    let args = parsed.unwrap_or_default();

    eprintln!("{}", render_preview(&tc.function.name, &args));

    let out = match tc.function.name.as_str() {
        "shell" => shell_tool(&args, policy),
        "read_file" => read_file_tool(&args),
        "write_file" => write_file_tool(&args, policy),
        "http_get" => http_get_tool(&args),
        "apply_patch" => apply_patch_tool(&args, policy),
        _ => Ok(json!({"error":"unknown tool"})),
    };

    out.unwrap_or_else(|err| json!({"error":err.to_string()}))
}

fn render_preview(name: &str, args: &HashMap<String, Value>) -> String {
    let raw = args.get("_raw").and_then(Value::as_str).unwrap_or("");
    match name {
        "shell" => format!("$ {}", preview(arg_or(args, "cmd", raw), 200)),
        "read_file" => format!("📄 {}", arg_or(args, "path", raw)),
        "write_file" => format!("✏️ {}", arg_or(args, "path", raw)),
        "http_get" => format!("🌐 {}", arg_or(args, "url", raw)),
        _ => format!("{} {}", name, preview(raw, 200)),
    }
}

fn arg_or<'a>(args: &'a HashMap<String, Value>, key: &str, default: &'a str) -> &'a str {
    args.get(key).and_then(Value::as_str).unwrap_or(default)
}

fn shell_tool(args: &HashMap<String, Value>, policy: &Policy) -> Result<Value> {
    let cmd = arg_or(args, "cmd", "");
    if cmd.is_empty() {
        return Err(anyhow!("missing cmd"));
    }
    for denied in &policy.deny {
        if cmd.contains(denied) {
            return Ok(json!({"error":"denied by policy"}));
        }
    }
    if !policy.allow.is_empty() && !policy.allow.iter().any(|allowed| cmd.contains(allowed)) {
        return Ok(json!({"error":"not allowed by policy"}));
    }
    if policy.dry_shell {
        return Ok(json!({"dry_run":true,"cmd":cmd}));
    }

    let mut command = Command::new("bash");
    command.arg("-lc").arg(cmd);
    if let Some(cwd) = &policy.cwd {
        command.current_dir(cwd);
    }
    let output = command.output().context("running shell command")?;
    Ok(json!({
        "returncode": output.status.code().unwrap_or(1),
        "stdout": tail(&String::from_utf8_lossy(&output.stdout), MAX_TOOL_OUTPUT_BYTES),
        "stderr": tail(&String::from_utf8_lossy(&output.stderr), MAX_TOOL_OUTPUT_BYTES),
    }))
}

fn read_file_tool(args: &HashMap<String, Value>) -> Result<Value> {
    let path = arg_or(args, "path", "");
    if path.is_empty() {
        return Err(anyhow!("missing path"));
    }
    let text = fs::read_to_string(path)?;
    if text.len() > MAX_READ_FILE_BYTES {
        return Ok(json!({"text": &text[..MAX_READ_FILE_BYTES], "truncated": true}));
    }
    Ok(json!({"text": text, "truncated": false}))
}

fn write_file_tool(args: &HashMap<String, Value>, policy: &Policy) -> Result<Value> {
    if policy.readonly {
        return Ok(json!({"error":"readonly session"}));
    }
    let path = arg_or(args, "path", "");
    if path.is_empty() {
        return Err(anyhow!("missing path"));
    }
    let text = arg_or(args, "text", "");
    fs::create_dir_all(dir_or_dot(path))?;
    fs::write(path, text)?;
    Ok(json!({"ok":true,"bytes":text.len()}))
}

fn http_get_tool(args: &HashMap<String, Value>) -> Result<Value> {
    let url = arg_or(args, "url", "");
    if url.is_empty() {
        return Err(anyhow!("missing url"));
    }
    let response = reqwest::blocking::get(url)?;
    let status = response.status().as_u16();
    let mut buf = String::new();
    response
        .take(MAX_TOOL_OUTPUT_BYTES as u64)
        .read_to_string(&mut buf)
        .ok();
    Ok(json!({"status": status, "body": buf}))
}

fn apply_patch_tool(args: &HashMap<String, Value>, policy: &Policy) -> Result<Value> {
    if policy.readonly {
        return Ok(json!({"error":"readonly session"}));
    }
    let patch = arg_or(args, "patch", "");
    if patch.is_empty() {
        return Err(anyhow!("missing patch"));
    }

    let mut cmd = Command::new("git");
    cmd.arg("apply").arg("--whitespace=nowarn").arg("-");
    if let Some(cwd) = &policy.cwd {
        cmd.current_dir(cwd);
    }
    let mut child = cmd
        .stdin(std::process::Stdio::piped())
        .stdout(std::process::Stdio::piped())
        .stderr(std::process::Stdio::piped())
        .spawn()?;
    use std::io::Write;
    if let Some(stdin) = child.stdin.as_mut() {
        stdin.write_all(patch.as_bytes())?;
    }
    let out = child.wait_with_output()?;
    if out.status.success() {
        return Ok(json!({"ok":true}));
    }
    Ok(json!({"error": String::from_utf8_lossy(&out.stderr)}))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn manifest_has_expected_tools() {
        let names: Vec<String> = tools_manifest()
            .into_iter()
            .map(|t| t.function.name)
            .collect();
        assert!(names.contains(&"shell".to_string()));
        assert!(names.contains(&"read_file".to_string()));
        assert!(names.contains(&"write_file".to_string()));
        assert!(names.contains(&"http_get".to_string()));
        assert!(names.contains(&"apply_patch".to_string()));
    }

    #[test]
    fn shell_policy_denies_command() {
        let mut args = HashMap::new();
        args.insert("cmd".to_string(), Value::String("rm -rf /".to_string()));
        let policy = Policy {
            deny: vec!["rm -rf".to_string()],
            ..Policy::default()
        };
        let out = shell_tool(&args, &policy).expect("tool result");
        assert_eq!(out["error"], "denied by policy");
    }
}
