const API_BASE = "http://localhost:8080";
let currentPlayerId = null;
let currentMatchId = null;
let lastRoundNumber = null;
let lastTurnNumber = null;
let lastTargetHPs = new Map(); // key: match_unit_id, value: current_hp

function appendLogEntry(text) {
  console.log("appendLogEntry:", text);
  const logEl = document.getElementById("battle-log");
  if (!logEl) return;
  const li = document.createElement("li");
  li.textContent = text;
  logEl.appendChild(li);
  logEl.scrollTop = logEl.scrollHeight;
}

async function fetchMatch() {
  if (!currentMatchId) {
    throw new Error("No match selected");
  }
  const res = await fetch(`${API_BASE}/matches/${currentMatchId}`);
  if (!res.ok) {
    throw new Error(`Failed to fetch match: ${res.status}`);
  }
  return res.json();
}

function renderMatch(data) {
  const matchInfoEl = document.getElementById("match-info");
  const sidesEl = document.getElementById("sides");

  const m = data.match;
  console.log("currentPlayerId:", currentPlayerId, "actor:", m.current_actor_player_id);

  const roundNumber = Math.floor((m.current_turn_number + 1) / 2);

  let turnInfo = "none";
  if (m.current_actor_player_id != null) {
    if (currentPlayerId && m.current_actor_player_id === currentPlayerId) {
      turnInfo = `you (player ${m.current_actor_player_id})`;
    } else {
      turnInfo = `player ${m.current_actor_player_id}`;
    }
  }

  matchInfoEl.textContent = `
State: ${m.state}
Round: ${roundNumber}
Turn: ${m.current_turn_number}
Current actor: ${turnInfo}
`.trim();

  sidesEl.innerHTML = "";

  // Optional: stable order
  data.sides.sort((a, b) => Number(a.player_id) - Number(b.player_id));

  const currentHPs = new Map();

  data.sides.forEach((side) => {
    const div = document.createElement("div");
    div.style.border = "1px solid #ccc";
    div.style.margin = "8px";
    div.style.padding = "4px";

    if (currentPlayerId != null && Number(side.player_id) === Number(currentPlayerId)) {
      div.style.backgroundColor = "#e6ffe6";
    } else {
      div.style.backgroundColor = "#f8f8f8";
    }

    const title = document.createElement("h3");
    title.textContent = `Player ${side.player_id} (active pos: ${side.active_position})`;
    div.appendChild(title);

    const list = document.createElement("ul");
    side.units.forEach((u) => {
      currentHPs.set(u.match_unit_id, u.current_hp);

      const li = document.createElement("li");
      li.textContent = `Unit ${u.unit_id} [match_unit ${u.match_unit_id}] pos=${u.position} HP=${u.current_hp}`;

      if (u.is_active && Array.isArray(u.moves) && u.moves.length > 0) {
        const movesContainer = document.createElement("div");
        movesContainer.textContent = "Moves: ";

        const isYourTurn =
          currentPlayerId != null &&
          m.current_actor_player_id != null &&
          Number(m.current_actor_player_id) === Number(currentPlayerId) &&
          Number(side.player_id) === Number(currentPlayerId);

        u.moves.forEach((mv) => {
          const btn = document.createElement("button");
          btn.textContent = `${mv.name} (id=${mv.id}, pow=${mv.power})`;
          btn.style.marginRight = "4px";

          if (!isYourTurn) {
            btn.disabled = true;
            btn.title = "Not your turn";
          } else {
            btn.addEventListener("click", () => {
              playTurnWithMove(mv.id);
            });
          }
          movesContainer.appendChild(btn);
        });

        li.appendChild(document.createElement("br"));
        li.appendChild(movesContainer);
      }

      list.appendChild(li);
    });

    div.appendChild(list);
    sidesEl.appendChild(div);
  });

  // Logging section
// Logging section
console.log("renderMatch: current_turn_number =", m.current_turn_number, "lastTurnNumber =", lastTurnNumber);

if (lastTurnNumber === null) {
  console.log("renderMatch: initializing lastTurnNumber");
  lastTurnNumber = m.current_turn_number;
  lastTargetHPs = currentHPs;
  return;
}

if (m.current_turn_number !== lastTurnNumber) {
  console.log("renderMatch: detected new turn", lastTurnNumber, "→", m.current_turn_number);

  const damages = [];

  currentHPs.forEach((hp, matchUnitId) => {
    const prev = lastTargetHPs.get(matchUnitId);
    if (prev != null && hp < prev) {
      damages.push({ matchUnitId, oldHp: prev, newHp: hp });
    }
  });

  console.log("renderMatch: damages computed =", damages);

  if (damages.length > 0) {
    damages.forEach((ev) => {
      appendLogEntry(
        `Turn ${lastTurnNumber}: match_unit ${ev.matchUnitId} took ${
          ev.oldHp - ev.newHp
        } damage (HP ${ev.oldHp} → ${ev.newHp})`
      );
    });
  } else {
    appendLogEntry(`Turn ${lastTurnNumber} resolved.`);
  }

  lastTurnNumber = m.current_turn_number;
  lastTargetHPs = currentHPs;
}
}

async function init() {
  document
    .getElementById("login-button")
    .addEventListener("click", login);

  document
    .getElementById("refresh-matches")
    .addEventListener("click", async () => {
      try {
        const matches = await fetchMyMatches();
        renderMatchList(matches);
      } catch (err) {
        document.getElementById("error").textContent = err.message;
      }
    });

  document
    .getElementById("create-match-button")
    .addEventListener("click", createMatch);

  document
    .getElementById("refresh-squads")
    .addEventListener("click", async () => {
      try {
        const squads = await fetchMySquads();
        renderSquadList(squads);
      } catch (err) {
        document.getElementById("error").textContent = err.message;
      }
    });

  document
    .getElementById("create-squad-button")
    .addEventListener("click", createSquad);

  try {
    const units = await fetchUnits();
    renderUnits(units);
  } catch (err) {
    document.getElementById("error").textContent = err.message;
  }

  startAutoRefresh();  // <= added
}

async function login() {
  const userEl = document.getElementById("login-username");
  const passEl = document.getElementById("login-password");
  const errorEl = document.getElementById("error");
  const statusEl = document.getElementById("status");

  errorEl.textContent = "";
  if (statusEl) statusEl.textContent = "";

  const username = userEl.value.trim();
  const password = passEl.value;

  if (!username || !password) {
    errorEl.textContent = "Username and password are required.";
    return;
  }

  try {
    const res = await fetch(`${API_BASE}/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password }),
    });

    if (!res.ok) {
      const text = await res.text();
      errorEl.textContent = `Login failed: ${res.status} ${text}`;
      return;
    }

    const data = await res.json();
    currentPlayerId = Number(data.player_id);
    document.getElementById("current-player").textContent =
      `${data.username} (id=${currentPlayerId})`;

    if (statusEl) statusEl.textContent = `Logged in as ${data.username}`;
  } catch (err) {
    errorEl.textContent = `Network error: ${err.message}`;
    return;
  }

  // Load match list after login
  try {
    const matches = await fetchMyMatches();
    renderMatchList(matches);
  } catch (err) {
    document.getElementById("error").textContent = err.message;
  }
}

async function playTurnWithMove(moveId) {
  const errorEl = document.getElementById("error");
  const statusEl = document.getElementById("status");
  errorEl.textContent = "";
  statusEl.textContent = "";

  if (!currentPlayerId) {
    errorEl.textContent = "You must log in first.";
    return;
  }

  if (!moveId) {
    errorEl.textContent = "Move ID is required.";
    return;
  }

  try {
    const res = await fetch(`${API_BASE}/matches/${currentMatchId}/turns`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Player-ID": String(currentPlayerId),
      },
      body: JSON.stringify({ move_id: moveId }),
    });

    if (!res.ok) {
      const text = await res.text();
      errorEl.textContent = `Error: ${res.status} ${text}`;
      return;
    }

    const data = await res.json();
    renderMatch(data);

    statusEl.textContent = "Turn played successfully.";
  } catch (err) {
    errorEl.textContent = `Network error: ${err.message}`;
  }
}

function renderMatchList(matches) {
  const listEl = document.getElementById("match-list");
  listEl.innerHTML = "";

  matches.forEach((m) => {
    const li = document.createElement("li");

    const btn = document.createElement("button");
    btn.textContent = `Match ${m.id} [${m.state}] P1=${m.player1_id}, P2=${m.player2_id}`;
    btn.addEventListener("click", async () => {
      currentMatchId = m.id;
      try {
        const matchData = await fetchMatch();
        renderMatch(matchData);
      } catch (err) {
        document.getElementById("error").textContent = err.message;
      }
    });

    li.appendChild(btn);
    listEl.appendChild(li);
  });
}

async function fetchMyMatches() {
  if (!currentPlayerId) {
    throw new Error("You must log in first to list matches.");
  }
  const res = await fetch(`${API_BASE}/me/matches`, {
    headers: {
      "X-Player-ID": String(currentPlayerId),
    },
  });
  if (!res.ok) {
    throw new Error(`Failed to fetch matches: ${res.status}`);
  }
  return res.json(); // array of MatchView
}

async function createMatch() {
  const errorEl = document.getElementById("error");
  const statusEl = document.getElementById("status");
  errorEl.textContent = "";
  if (statusEl) statusEl.textContent = "";

  if (!currentPlayerId) {
    errorEl.textContent = "You must log in first.";
    return;
  }

  const oppEl = document.getElementById("create-opponent-id");
  const squad1El = document.getElementById("create-squad1-id");
  const squad2El = document.getElementById("create-squad2-id");

  const opponentId = Number(oppEl.value);
  const squad1Id = Number(squad1El.value);
  const squad2Id = Number(squad2El.value);

  if (!opponentId || !squad1Id || !squad2Id) {
    errorEl.textContent = "Opponent ID and both squad IDs are required.";
    return;
  }

  try {
    const res = await fetch(`${API_BASE}/matches`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Player-ID": String(currentPlayerId),
      },
      body: JSON.stringify({
        opponent_player_id: opponentId,
        player1_squad_id: squad1Id,
        player2_squad_id: squad2Id,
      }),
    });

    if (!res.ok) {
      const text = await res.text();
      errorEl.textContent = `Create match failed: ${res.status} ${text}`;
      return;
    }

    const data = await res.json();
    currentMatchId = data.match.id; // from MatchResponse
    renderMatch(data);

    if (statusEl) statusEl.textContent = `Created and selected match ${currentMatchId}`;

    // Refresh match list
    const matches = await fetchMyMatches();
    renderMatchList(matches);
  } catch (err) {
    errorEl.textContent = `Network error: ${err.message}`;
  }
}

async function fetchUnits() {
  const res = await fetch(`${API_BASE}/units`);
  if (!res.ok) {
    throw new Error(`Failed to fetch units: ${res.status}`);
  }
  return res.json(); // [{id, name, ...}]
}

function renderUnits(units) {
  const listEl = document.getElementById("unit-list");
  listEl.innerHTML = "";
  units.forEach((u) => {
    const li = document.createElement("li");
    li.textContent = `ID=${u.id} Name=${u.name} HP=${u.base_hp} ATK=${u.base_attack} SPD=${u.base_speed}`;
    listEl.appendChild(li);
  });
}

async function fetchMySquads() {
  if (!currentPlayerId) {
    throw new Error("You must log in first to list squads.");
  }
  const res = await fetch(`${API_BASE}/me/squads`, {
    headers: {
      "X-Player-ID": String(currentPlayerId),
    },
  });
  if (!res.ok) {
    throw new Error(`Failed to fetch squads: ${res.status}`);
  }
  return res.json(); // [{id, name, units:[...]}]
}

function renderSquadList(squads) {
  const listEl = document.getElementById("squad-list");
  listEl.innerHTML = "";

  squads.forEach((sq) => {
    const li = document.createElement("li");
    li.textContent = `Squad ${sq.id} "${sq.name}" units=${sq.units.join(", ")}`;
    listEl.appendChild(li);
  });
}

async function createSquad() {
  const errorEl = document.getElementById("error");
  const statusEl = document.getElementById("status");
  errorEl.textContent = "";
  if (statusEl) statusEl.textContent = "";

  if (!currentPlayerId) {
    errorEl.textContent = "You must log in first.";
    return;
  }

  const nameEl = document.getElementById("squad-name");
  const unitsEl = document.getElementById("squad-unit-ids");

  const name = nameEl.value.trim();
  const unitsRaw = unitsEl.value.trim();

  if (!name || !unitsRaw) {
    errorEl.textContent = "Squad name and unit IDs are required.";
    return;
  }

  // Parse comma-separated unit IDs
  const unitIds = unitsRaw
    .split(",")
    .map((s) => s.trim())
    .filter((s) => s.length > 0)
    .map((s) => Number(s))
    .filter((n) => !Number.isNaN(n));

  if (unitIds.length === 0) {
    errorEl.textContent = "Please provide at least one valid unit ID.";
    return;
  }

  try {
    const res = await fetch(`${API_BASE}/me/squads`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Player-ID": String(currentPlayerId),
      },
      body: JSON.stringify({
        name: name,
        unit_ids: unitIds,
      }),
    });

    if (!res.ok) {
      const text = await res.text();
      errorEl.textContent = `Create squad failed: ${res.status} ${text}`;
      return;
    }

    if (statusEl) statusEl.textContent = `Created squad "${name}"`;

    // Refresh squads list
    const squads = await fetchMySquads();
    renderSquadList(squads);
  } catch (err) {
    errorEl.textContent = `Network error: ${err.message}`;
  }
}

function startAutoRefresh() {
  // refresh every 5 seconds
  setInterval(async () => {
    if (!currentPlayerId) return;          // not logged in, skip
    try {
      const matches = await fetchMyMatches();
      renderMatchList(matches);

      if (currentMatchId) {
        const matchData = await fetchMatch();
        renderMatch(matchData);
      }
    } catch (err) {
      // optional: comment this out if too noisy
      console.log("auto-refresh error:", err.message);
    }
  }, 5000);
}

init();