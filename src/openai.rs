use anyhow::{anyhow, Result};
use reqwest::blocking::Client;
use reqwest::header::{AUTHORIZATION, CONTENT_TYPE};
use serde_json::json;

use crate::tools_runtime::{execute_tool_call, tools_manifest};
use crate::types::{ChatRequest, ChatResponse, Message, Policy, ToolCall};

const MAX_CHAT_TURNS: usize = 100;

pub fn chat_session(
    model: &str,
    initial_messages: Vec<Message>,
    policy: &Policy,
) -> Result<String> {
    let mut messages = initial_messages;
    let client = Client::new();
    let tools = tools_manifest();

    for _ in 0..MAX_CHAT_TURNS {
        let req = ChatRequest {
            model: model.to_string(),
            messages: messages.clone(),
            tools: tools.clone(),
            tool_choice: "auto".to_string(),
        };

        let resp = chat_once(&client, &req)?;
        let choice = resp
            .choices
            .into_iter()
            .next()
            .ok_or_else(|| anyhow!("no choices"))?;
        let assistant = choice.message;
        let tool_calls = assistant.tool_calls.clone();
        messages.push(assistant.clone());

        if !tool_calls.is_empty() {
            for tc in tool_calls {
                messages.push(tool_output_message(&tc, execute_tool_call(&tc, policy)));
            }
            continue;
        }

        return Ok(assistant.content.unwrap_or_default());
    }

    Err(anyhow!("max turns reached"))
}

fn chat_once(client: &Client, req: &ChatRequest) -> Result<ChatResponse> {
    let base =
        std::env::var("OPENAI_BASE_URL").unwrap_or_else(|_| "https://api.openai.com".to_string());
    let key = std::env::var("OPENAI_API_KEY").map_err(|_| anyhow!("OPENAI_API_KEY not set"))?;

    let request = client
        .post(format!(
            "{}/v1/chat/completions",
            base.trim_end_matches('/')
        ))
        .header(AUTHORIZATION, format!("Bearer {key}"))
        .header(CONTENT_TYPE, "application/json")
        .json(req);

    if std::env::var("DEBUG").ok().as_deref() == Some("1") {
        eprintln!(
            "--- DEBUG REQUEST ---\n{}",
            serde_json::to_string_pretty(req)?
        );
    }

    let response = request.send()?;
    let status = response.status();
    let text = response.text()?;

    if std::env::var("DEBUG").ok().as_deref() == Some("1") {
        eprintln!("--- DEBUG RESPONSE ({status}) ---\n{text}\n---");
    }

    if !status.is_success() {
        return Err(anyhow!("API {}: {}", status.as_u16(), text));
    }

    Ok(serde_json::from_str(&text)?)
}

fn tool_output_message(tc: &ToolCall, output: serde_json::Value) -> Message {
    Message {
        role: "tool".to_string(),
        content: Some(json!(output).to_string()),
        name: Some(tc.function.name.clone()),
        tool_call_id: Some(tc.id.clone()),
        tool_calls: vec![],
    }
}
