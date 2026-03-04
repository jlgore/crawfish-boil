package gateway

// spaHTML is an embedded fake OpenClaw SPA shell mimicking the real Vite+Lit application.
const spaHTML = `<!doctype html>
<html lang="en" data-theme="dark">
<head>
  <meta charset="UTF-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
  <title>OpenClaw — AI Agent Platform</title>
  <style>
    :root {
      --bg: #12141a;
      --bg-secondary: #1a1d26;
      --accent: #ff5c5c;
      --accent-hover: #ff7a7a;
      --text: #e4e6ed;
      --text-muted: #8b8fa3;
      --border: #2a2d3a;
      --font: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
    }
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      background: var(--bg);
      color: var(--text);
      font-family: var(--font);
      height: 100vh;
      display: flex;
      flex-direction: column;
    }
    header {
      display: flex;
      align-items: center;
      padding: 12px 20px;
      border-bottom: 1px solid var(--border);
      background: var(--bg-secondary);
    }
    header .logo {
      display: flex;
      align-items: center;
      gap: 10px;
      font-weight: 700;
      font-size: 16px;
    }
    header .logo svg { width: 28px; height: 28px; }
    header .logo .version {
      font-size: 11px;
      color: var(--text-muted);
      background: var(--bg);
      padding: 2px 8px;
      border-radius: 10px;
    }
    header nav {
      margin-left: auto;
      display: flex;
      gap: 16px;
      align-items: center;
    }
    header nav a {
      color: var(--text-muted);
      text-decoration: none;
      font-size: 13px;
      transition: color 0.15s;
    }
    header nav a:hover { color: var(--text); }
    .avatar {
      width: 30px;
      height: 30px;
      border-radius: 50%;
      background: var(--accent);
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 13px;
      font-weight: 600;
      color: #fff;
    }
    main {
      flex: 1;
      display: flex;
      align-items: center;
      justify-content: center;
      flex-direction: column;
      gap: 16px;
    }
    .spinner {
      width: 32px; height: 32px;
      border: 3px solid var(--border);
      border-top-color: var(--accent);
      border-radius: 50%;
      animation: spin 0.8s linear infinite;
    }
    @keyframes spin { to { transform: rotate(360deg); } }
    .status { color: var(--text-muted); font-size: 14px; }
    footer {
      padding: 8px 20px;
      border-top: 1px solid var(--border);
      font-size: 11px;
      color: var(--text-muted);
      text-align: center;
    }
  </style>
</head>
<body>
  <header>
    <div class="logo">
      <svg viewBox="0 0 28 28" fill="none" xmlns="http://www.w3.org/2000/svg">
        <rect width="28" height="28" rx="6" fill="#ff5c5c"/>
        <path d="M8 10C8 8.89543 8.89543 8 10 8H18C19.1046 8 20 8.89543 20 10V14L14 20H10C8.89543 20 8 19.1046 8 18V10Z" fill="#fff" fill-opacity="0.9"/>
        <circle cx="12" cy="13" r="1.5" fill="#ff5c5c"/>
        <circle cx="17" cy="13" r="1.5" fill="#ff5c5c"/>
      </svg>
      OpenClaw
      <span class="version">v0.14.2</span>
    </div>
    <nav>
      <a href="/api/channels">Channels</a>
      <a href="/__openclaw__/canvas">Canvas</a>
      <a href="#">Docs</a>
      <div class="avatar">A</div>
    </nav>
  </header>
  <main>
    <openclaw-app>
      <div class="spinner"></div>
      <div class="status">Connecting to agent…</div>
    </openclaw-app>
  </main>
  <footer>OpenClaw v0.14.2 · Protocol v3 · Agent runtime ready</footer>
  <script type="module">
    // OpenClaw SPA bootstrap (Lit web components)
    const ws = new WebSocket((location.protocol === 'https:' ? 'wss://' : 'ws://') + location.host + '/');
    ws.onopen = () => document.querySelector('.status').textContent = 'Connected — authenticating…';
    ws.onclose = () => document.querySelector('.status').textContent = 'Disconnected';
  </script>
</body>
</html>`
