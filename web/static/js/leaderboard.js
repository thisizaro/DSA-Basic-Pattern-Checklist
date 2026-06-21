/**
 * leaderboard.js — powers leaderboard.html. Fully public: works for both
 * signed-in and signed-out visitors, since the leaderboard and platform
 * stats are open by design (the project's "flex" page for open-sourcing).
 */

const contentEl = document.getElementById("content");

init();

async function init() {
  let me = null;
  try {
    me = await api.auth.me();
  } catch {
    me = null; // logged-out visitor — perfectly fine here
  }

  renderNav("leaderboard", me && me.account_status === "active" ? me : null);

  try {
    const [stats, entries] = await Promise.all([
      api.leaderboard.getStats(),
      api.leaderboard.getAll(),
    ]);
    render(stats, entries || []);
  } catch (err) {
    contentEl.innerHTML = `<p class="state-message">Couldn't load the leaderboard: ${escapeHtml(err.message)}</p>`;
  }
}

function render(stats, entries) {
  contentEl.innerHTML = `
    ${renderStatsGrid(stats)}
    ${renderSignupChart(stats.signups_by_day)}
    ${renderTable(entries)}
  `;
}

function renderStatsGrid(stats) {
  const tiles = [
    { num: stats.total_users, label: "Active users" },
    { num: stats.total_patterns_solved, label: "Patterns solved (all-time)" },
    { num: stats.total_steps_completed, label: "Checklist steps checked" },
    { num: stats.total_colleges, label: "Colleges represented" },
  ];

  return `
    <div class="stats-grid">
      ${tiles
        .map(
          (t) => `
        <div class="stat-tile">
          <div class="num">${t.num.toLocaleString()}</div>
          <div class="label">${escapeHtml(t.label)}</div>
        </div>
      `
        )
        .join("")}
    </div>
  `;
}

function renderSignupChart(signupsByDay) {
  if (!signupsByDay || signupsByDay.length === 0) return "";

  const max = Math.max(...signupsByDay.map((d) => d.count), 1);

  const bars = signupsByDay
    .map((d) => {
      const heightPct = Math.max((d.count / max) * 100, 3);
      return `<div class="signup-bar" style="height:${heightPct}%" title="${escapeAttr(d.date)}: ${d.count} signup${d.count === 1 ? "" : "s"}"></div>`;
    })
    .join("");

  return `
    <div class="profile-section">
      <h2>Signups, last 30 days</h2>
      <div class="signup-chart">${bars}</div>
    </div>
  `;
}

function renderTable(entries) {
  if (entries.length === 0) {
    return `<p class="state-message">No one's on the board yet — be the first to check off a pattern.</p>`;
  }

  const rows = entries
    .map((e, i) => {
      const rank = i + 1;
      const isTopThree = rank <= 3;

      const avatar = e.avatar_url
        ? `<img class="leaderboard-avatar" src="${escapeAttr(e.avatar_url)}" alt="" />`
        : `<div class="leaderboard-avatar"></div>`;

      const nameCell = e.is_anonymous
        ? `<span class="leaderboard-username">Anonymous</span>`
        : `<a href="/u/${encodeURIComponent(e.username)}"><span class="leaderboard-username">${escapeHtml(e.display_name)}</span></a>`;

      return `
        <tr class="leaderboard-row${isTopThree ? " top-three" : ""}">
          <td class="col-num rank">${rank}</td>
          <td>
            <div class="leaderboard-user">
              ${avatar}
              <div>
                ${nameCell}
                ${e.college_name ? `<div class="leaderboard-college">${escapeHtml(e.college_name)}</div>` : ""}
              </div>
            </div>
          </td>
          <td class="col-solved">${e.patterns_solved}</td>
          <td class="col-steps">${e.steps_completed}</td>
        </tr>
      `;
    })
    .join("");

  return `
    <table class="leaderboard-table">
      <thead>
        <tr>
          <th class="col-num">#</th>
          <th>User</th>
          <th class="col-solved" style="text-align:center;">Solved</th>
          <th class="col-steps" style="text-align:center;">Steps</th>
        </tr>
      </thead>
      <tbody>${rows}</tbody>
    </table>
  `;
}

function escapeAttr(str) {
  return escapeHtml(str).replace(/"/g, "&quot;");
}
