package web

const indexHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Jorin Web</title>
  <style>
    body { font-family: system-ui, sans-serif; margin: 0; background: #0f172a; color: #e2e8f0; }
    .wrap { max-width: 980px; margin: 0 auto; padding: 20px; }
    h1 { margin-top: 0; }
    .chat { border: 1px solid #334155; border-radius: 8px; background: #111827; min-height: 360px; padding: 12px; overflow-y: auto; }
    .msg { margin: 10px 0; white-space: pre-wrap; }
    .user { color: #93c5fd; }
    .assistant { color: #86efac; }
    form { display: flex; gap: 8px; margin-top: 12px; }
    input[type="text"] { flex: 1; padding: 10px; border-radius: 8px; border: 1px solid #334155; background: #0b1220; color: #e2e8f0; }
    button { padding: 10px 14px; border-radius: 8px; border: 1px solid #334155; background: #1d4ed8; color: white; cursor: pointer; }
  </style>
</head>
<body>
  <div class="wrap">
    <h1>Jorin Web Service</h1>
    <p>Ask Jorin from your browser. Messages are kept in this tab only.</p>
    <div id="chat" class="chat" aria-live="polite"></div>
    <form id="chatForm">
      <input id="prompt" type="text" placeholder="Ask Jorin to refactor, test, explain..." required />
      <button type="submit">Send</button>
    </form>
  </div>
  <script>
    const chat = document.getElementById('chat');
    const form = document.getElementById('chatForm');
    const promptInput = document.getElementById('prompt');
    const messages = [];

    function append(role, content) {
      const div = document.createElement('div');
      div.className = 'msg ' + role;
      div.textContent = (role === 'user' ? 'You: ' : 'Jorin: ') + content;
      chat.appendChild(div);
      chat.scrollTop = chat.scrollHeight;
    }

    form.addEventListener('submit', async (event) => {
      event.preventDefault();
      const prompt = promptInput.value.trim();
      if (!prompt) return;

      messages.push({ role: 'user', content: prompt });
      append('user', prompt);
      promptInput.value = '';

      const response = await fetch('/api/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ messages })
      });
      if (!response.ok) {
        const err = await response.text();
        append('assistant', 'Error: ' + err);
        return;
      }

      const data = await response.json();
      messages.push({ role: 'assistant', content: data.reply });
      append('assistant', data.reply);
    });
  </script>
</body>
</html>
`
