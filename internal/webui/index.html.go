package webui

const indexHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>PokerShow</title>
<link rel="preconnect" href="https://fonts.googleapis.com" />
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
<link href="https://fonts.googleapis.com/css2?family=Figtree:wght@400;600;700&family=Newsreader:opsz,wght@6..72,500;6..72,700&display=swap" rel="stylesheet" />
<style>
  :root {
    --felt-deep: #0c2e24;
    --felt: #145c45;
    --felt-light: #1f7a5a;
    --ink: #e8f2ec;
    --muted: #a8c4b6;
    --brass: #d4b56a;
    --brass-dim: #9a8040;
    --danger: #e07a6a;
    --card-face: #f7f3ea;
    --card-red: #c23b2e;
    --card-black: #1a1f1c;
    --panel: rgba(6, 22, 18, 0.72);
    --line: rgba(212, 181, 106, 0.28);
  }
  * { box-sizing: border-box; }
  html, body {
    margin: 0; min-height: 100%;
    color: var(--ink);
    font-family: "Figtree", sans-serif;
    background:
      radial-gradient(1200px 600px at 50% -10%, rgba(212,181,106,0.18), transparent 55%),
      radial-gradient(800px 500px at 80% 100%, rgba(20,92,69,0.55), transparent 50%),
      linear-gradient(165deg, #071912 0%, var(--felt-deep) 40%, #0a241c 100%);
  }
  body::before {
    content: "";
    position: fixed; inset: 0; pointer-events: none; opacity: 0.07;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='120' height='120'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='.8' numOctaves='2' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='120' height='120' filter='url(%23n)' opacity='.55'/%3E%3C/svg%3E");
  }
  .wrap {
    max-width: 960px;
    margin: 0 auto;
    padding: 28px 20px 48px;
    position: relative;
  }
  header {
    display: flex; flex-wrap: wrap; align-items: end; justify-content: space-between;
    gap: 16px; margin-bottom: 22px;
    animation: rise 0.6s ease both;
  }
  .brand {
    font-family: "Newsreader", serif;
    font-size: clamp(2.4rem, 6vw, 3.4rem);
    font-weight: 700;
    letter-spacing: -0.02em;
    line-height: 0.95;
    margin: 0;
  }
  .brand span { color: var(--brass); }
  .tagline {
    margin: 8px 0 0;
    color: var(--muted);
    font-size: 0.95rem;
    max-width: 28ch;
  }
  .controls {
    display: grid;
    grid-template-columns: repeat(4, minmax(64px, 1fr)) auto;
    gap: 8px;
    align-items: end;
    background: var(--panel);
    border: 1px solid var(--line);
    padding: 12px;
    backdrop-filter: blur(8px);
  }
  label {
    display: flex; flex-direction: column; gap: 4px;
    font-size: 0.72rem; letter-spacing: 0.06em; text-transform: uppercase; color: var(--muted);
  }
  input {
    width: 100%;
    background: rgba(0,0,0,0.35);
    border: 1px solid rgba(255,255,255,0.08);
    color: var(--ink);
    padding: 8px 10px;
    font: inherit;
    border-radius: 2px;
  }
  button {
    font: inherit; cursor: pointer;
    border: 1px solid var(--line);
    background: linear-gradient(180deg, #2a6b52, #1a4a39);
    color: var(--ink);
    padding: 10px 14px;
    border-radius: 2px;
    transition: transform 0.15s ease, border-color 0.15s ease, filter 0.15s ease;
  }
  button:hover:not(:disabled) {
    border-color: var(--brass);
    filter: brightness(1.08);
    transform: translateY(-1px);
  }
  button:disabled {
    opacity: 0.35; cursor: not-allowed; transform: none;
  }
  button.ghost {
    background: transparent;
  }
  button.danger {
    background: linear-gradient(180deg, #8a3a32, #5c241f);
    border-color: rgba(224,122,106,0.45);
  }
  button.primary {
    background: linear-gradient(180deg, #c9a84a, #8f7028);
    color: #1a1408;
    border-color: rgba(255,255,255,0.2);
    font-weight: 700;
  }
  .table {
    position: relative;
    margin-top: 8px;
    padding: 28px 22px 24px;
    border-radius: 999px / 120px;
    background:
      radial-gradient(ellipse at center, var(--felt-light) 0%, var(--felt) 45%, #0e3d2e 100%);
    box-shadow:
      inset 0 0 0 3px rgba(212,181,106,0.35),
      inset 0 0 60px rgba(0,0,0,0.35),
      0 24px 50px rgba(0,0,0,0.45);
    min-height: 360px;
    animation: tableIn 0.7s cubic-bezier(.2,.8,.2,1) both;
  }
  .meta {
    display: flex; flex-wrap: wrap; gap: 10px 18px; justify-content: center;
    margin-bottom: 18px;
    font-size: 0.9rem; color: var(--muted);
  }
  .meta strong { color: var(--brass); font-weight: 600; }
  .board {
    display: flex; justify-content: center; gap: 10px; min-height: 96px;
    margin: 10px 0 28px;
  }
  .seats {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 18px;
    max-width: 640px;
    margin: 0 auto;
  }
  .seat {
    text-align: center;
    padding: 14px 12px;
    border: 1px solid transparent;
    transition: border-color 0.2s ease, background 0.2s ease;
  }
  .seat.active {
    border-color: var(--brass);
    background: rgba(0,0,0,0.22);
    animation: pulse 1.6s ease-in-out infinite;
  }
  .seat.folded { opacity: 0.45; }
  .seat .name {
    font-family: "Newsreader", serif;
    font-size: 1.25rem;
    margin-bottom: 4px;
  }
  .seat .stack { color: var(--muted); font-size: 0.85rem; margin-bottom: 10px; }
  .hole { display: flex; justify-content: center; gap: 8px; min-height: 84px; }
  .card {
    width: 58px; height: 84px;
    border-radius: 6px;
    background: var(--card-face);
    color: var(--card-black);
    display: flex; flex-direction: column; align-items: center; justify-content: center;
    font-weight: 700;
    box-shadow: 0 6px 14px rgba(0,0,0,0.35);
    animation: deal 0.35s ease both;
  }
  .card.red { color: var(--card-red); }
  .card .rank { font-size: 1.25rem; line-height: 1; }
  .card .suit { font-size: 1.1rem; }
  .card.back {
    background: repeating-linear-gradient(
      -45deg, #1f3d55, #1f3d55 6px, #254a66 6px, #254a66 12px);
    color: transparent;
  }
  .card.empty {
    background: rgba(255,255,255,0.06);
    border: 1px dashed rgba(255,255,255,0.15);
    box-shadow: none;
  }
  .actions {
    display: flex; flex-wrap: wrap; gap: 10px; justify-content: center;
    margin: 22px 0 8px;
    animation: rise 0.5s ease 0.15s both;
  }
  .raise-box {
    display: flex; align-items: center; gap: 8px;
    background: var(--panel);
    border: 1px solid var(--line);
    padding: 6px 8px 6px 12px;
  }
  .raise-box input { width: 88px; }
  .status {
    text-align: center;
    min-height: 1.4em;
    color: var(--brass);
    font-size: 0.95rem;
  }
  .status.err { color: var(--danger); }
  .log-wrap {
    margin-top: 22px;
    background: var(--panel);
    border: 1px solid var(--line);
    padding: 14px 16px;
  }
  .log-wrap h2 {
    margin: 0 0 10px;
    font-family: "Newsreader", serif;
    font-size: 1.15rem;
    font-weight: 500;
    color: var(--brass);
  }
  .log {
    max-height: 200px; overflow: auto;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
    font-size: 0.8rem;
    line-height: 1.55;
    color: var(--muted);
    white-space: pre-wrap;
  }
  @keyframes rise {
    from { opacity: 0; transform: translateY(10px); }
    to { opacity: 1; transform: none; }
  }
  @keyframes tableIn {
    from { opacity: 0; transform: scale(0.97); }
    to { opacity: 1; transform: none; }
  }
  @keyframes deal {
    from { opacity: 0; transform: translateY(-12px) rotate(-4deg); }
    to { opacity: 1; transform: none; }
  }
  @keyframes pulse {
    0%, 100% { box-shadow: 0 0 0 0 rgba(212,181,106,0.0); }
    50% { box-shadow: 0 0 0 4px rgba(212,181,106,0.18); }
  }
  @media (max-width: 720px) {
    .controls { grid-template-columns: 1fr 1fr; }
    .controls button { grid-column: 1 / -1; }
    .table { border-radius: 28px; min-height: 0; }
    .seats { grid-template-columns: 1fr; }
  }
</style>
</head>
<body>
  <div class="wrap">
    <header>
      <div>
        <h1 class="brand">Poker<span>Show</span></h1>
        <p class="tagline">本机德州调试桌：点按钮出牌，种子可复现。</p>
      </div>
      <form class="controls" id="newForm">
        <label>Seed <input name="seed" type="number" value="42" /></label>
        <label>Stack <input name="stack" type="number" value="1000" /></label>
        <label>SB <input name="sb" type="number" value="5" /></label>
        <label>BB <input name="bb" type="number" value="10" /></label>
        <button type="submit" class="primary">新开局</button>
      </form>
    </header>

    <section class="table" id="table">
      <div class="meta">
        <span>手牌 <strong id="handNo">-</strong></span>
        <span>街道 <strong id="street">-</strong></span>
        <span>底池 <strong id="pot">-</strong></span>
        <span>当前注 <strong id="currentBet">-</strong></span>
        <span>盲注 <strong id="blinds">-</strong></span>
      </div>
      <div class="board" id="board"></div>
      <div class="seats" id="seats"></div>
      <div class="actions" id="actions"></div>
      <div class="status" id="status"></div>
    </section>

    <section class="log-wrap">
      <h2>本手日志</h2>
      <div class="log" id="log"></div>
    </section>
  </div>
<script>
const suitMap = { c: "♣", d: "♦", h: "♥", s: "♠" };
const redSuits = new Set(["d", "h"]);

function parseCard(code) {
  if (!code) return null;
  const suit = code.slice(-1).toLowerCase();
  const rank = code.slice(0, -1).toUpperCase();
  return { rank, suit, red: redSuits.has(suit), label: suitMap[suit] || suit };
}

function cardEl(code, opts = {}) {
  const el = document.createElement("div");
  if (!code) {
    el.className = "card " + (opts.back ? "back" : "empty");
    return el;
  }
  const c = parseCard(code);
  el.className = "card" + (c.red ? " red" : "");
  el.innerHTML = '<div class="rank">' + c.rank + '</div><div class="suit">' + c.label + '</div>';
  el.style.animationDelay = (opts.delay || 0) + "ms";
  return el;
}

async function api(path, body) {
  const opts = body ? {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  } : {};
  const res = await fetch(path, opts);
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    const err = data.error || res.statusText;
    throw new Error(err);
  }
  return data;
}

function render(state) {
  document.getElementById("handNo").textContent = "#" + state.handNo;
  document.getElementById("street").textContent = state.street;
  document.getElementById("pot").textContent = state.pot;
  document.getElementById("currentBet").textContent = state.currentBet;
  document.getElementById("blinds").textContent = state.smallBlind + "/" + state.bigBlind;

  const board = document.getElementById("board");
  board.innerHTML = "";
  const boardCards = state.board || [];
  for (let i = 0; i < 5; i++) {
    board.appendChild(cardEl(boardCards[i], { delay: i * 40 }));
  }

  const seats = document.getElementById("seats");
  seats.innerHTML = "";
  (state.players || []).forEach((p, idx) => {
    const seat = document.createElement("div");
    seat.className = "seat" + (p.toAct ? " active" : "") + (p.folded ? " folded" : "");
    const flags = [];
    if (p.allIn) flags.push("全下");
    if (p.folded) flags.push("弃牌");
    const hole = document.createElement("div");
    hole.className = "hole";
    if (p.hole && p.hole.length) {
      p.hole.forEach((c, i) => hole.appendChild(cardEl(c, { delay: i * 50 })));
    } else if (!p.human && !state.finished) {
      hole.appendChild(cardEl(null, { back: true }));
      hole.appendChild(cardEl(null, { back: true }));
    } else {
      hole.appendChild(cardEl(null));
      hole.appendChild(cardEl(null));
    }
    seat.innerHTML =
      '<div class="name">' + p.name + (p.toAct ? " · 行动中" : "") + '</div>' +
      '<div class="stack">筹码 ' + p.stack + ' · 本街 ' + p.bet +
      (flags.length ? " · " + flags.join(" ") : "") + '</div>';
    seat.appendChild(hole);
    seats.appendChild(seat);
  });

  const actions = document.getElementById("actions");
  actions.innerHTML = "";
  const status = document.getElementById("status");
  status.className = "status";
  if (state.error) {
    status.classList.add("err");
    status.textContent = state.error;
  } else if (state.finished) {
    status.textContent = "本手结束";
    const next = document.createElement("button");
    next.className = "primary";
    next.textContent = "下一手";
    next.onclick = async () => {
      try { render(await api("/api/next", {})); }
      catch (e) { status.classList.add("err"); status.textContent = e.message; }
    };
    actions.appendChild(next);
  } else if (state.yourTurn) {
    status.textContent = "轮到你了";
    const legal = state.legal || [];
    const kinds = new Set(legal.map(a => a.kind));
    const byKind = Object.fromEntries(legal.map(a => [a.kind, a]));

    if (kinds.has("fold")) {
      const b = document.createElement("button");
      b.className = "danger";
      b.textContent = "弃牌";
      b.onclick = () => doAct({ kind: "fold" });
      actions.appendChild(b);
    }
    if (kinds.has("check")) {
      const b = document.createElement("button");
      b.textContent = "过牌";
      b.onclick = () => doAct({ kind: "check" });
      actions.appendChild(b);
    }
    if (kinds.has("call")) {
      const b = document.createElement("button");
      b.className = "primary";
      b.textContent = byKind.call.label || "跟注";
      b.onclick = () => doAct({ kind: "call" });
      actions.appendChild(b);
    }
    if (kinds.has("raise")) {
      const box = document.createElement("div");
      box.className = "raise-box";
      const min = byKind.raise.minRaiseTo || (state.currentBet + state.minRaise);
      const input = document.createElement("input");
      input.type = "number";
      input.value = min;
      input.min = min;
      const b = document.createElement("button");
      b.className = "primary";
      b.textContent = "加注到";
      b.onclick = () => doAct({ kind: "raise", amount: Number(input.value) });
      box.appendChild(document.createTextNode("≥" + min));
      box.appendChild(input);
      box.appendChild(b);
      actions.appendChild(box);
    }
    if (kinds.has("allin")) {
      const b = document.createElement("button");
      b.className = "ghost";
      b.textContent = "全下";
      b.onclick = () => doAct({ kind: "allin" });
      actions.appendChild(b);
    }
  } else {
    status.textContent = "等待中…";
  }

  document.getElementById("log").textContent = (state.log || []).join("\n");
  const form = document.getElementById("newForm");
  form.seed.value = state.seed;
  form.stack.value = state.stackStart;
  form.sb.value = state.smallBlind;
  form.bb.value = state.bigBlind;
}

async function doAct(body) {
  const status = document.getElementById("status");
  try {
    render(await api("/api/action", body));
  } catch (e) {
    status.className = "status err";
    status.textContent = e.message;
    try { render(await api("/api/state")); } catch (_) {}
  }
}

document.getElementById("newForm").addEventListener("submit", async (e) => {
  e.preventDefault();
  const fd = new FormData(e.target);
  try {
    render(await api("/api/new", {
      seed: Number(fd.get("seed")),
      stack: Number(fd.get("stack")),
      smallBlind: Number(fd.get("sb")),
      bigBlind: Number(fd.get("bb")),
    }));
  } catch (err) {
    const status = document.getElementById("status");
    status.className = "status err";
    status.textContent = err.message;
  }
});

api("/api/state").then(render).catch((e) => {
  document.getElementById("status").className = "status err";
  document.getElementById("status").textContent = e.message;
});
</script>
</body>
</html>
`
