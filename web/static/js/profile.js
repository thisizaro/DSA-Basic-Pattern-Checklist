/**
 * profile.js — powers /u/<username>. If the viewer is signed in and the
 * username in the URL matches their own, renders an editable profile
 * (username, LeetCode link, anonymous toggle). Otherwise renders the
 * public read-only view, which respects the profile owner's anonymity
 * setting.
 */

const contentEl = document.getElementById("content");

// /u/<username> is served by the Go router for this exact path — pull the
// username back out of the URL here.
const pathParts = window.location.pathname.split("/").filter(Boolean);
const requestedUsername = decodeURIComponent(pathParts[pathParts.length - 1] || "");

init();

async function init() {
  let me = null;
  try {
    me = await api.auth.me();
  } catch {
    // Not signed in — that's fine, public profiles don't require auth.
    me = null;
  }

  renderNav("profile", me && me.account_status === "active" ? me : null);

  const isOwnProfile = me && me.username === requestedUsername;

  if (isOwnProfile) {
    await renderEditableProfile(me);
  } else {
    await renderPublicProfile(requestedUsername);
  }
}

/* ── Public (read-only) view ──────────────────────────────────────── */

async function renderPublicProfile(username) {
  let profile;
  try {
    profile = await api.profile.getPublic(username);
  } catch (err) {
    contentEl.innerHTML = `<p class="state-message">${err.status === 404 ? "No user found at this profile." : escapeHtml(err.message)}</p>`;
    return;
  }

  const joined = new Date(profile.joined_at).toLocaleDateString(undefined, {
    year: "numeric",
    month: "long",
  });

  contentEl.innerHTML = `
    ${renderProfileHeader(profile)}
    ${renderStatsGrid(profile.summary)}
    <div class="profile-section">
      <h2>Details</h2>
      <div class="external-link-row">Joined ${escapeHtml(joined)}</div>
      ${profile.leetcode_url ? `<div class="external-link-row">LeetCode — <a href="${escapeAttr(profile.leetcode_url)}" target="_blank" rel="noopener noreferrer">${escapeHtml(profile.leetcode_url)}</a></div>` : ""}
    </div>
  `;
}

function renderProfileHeader(profile, alwaysShowUsername) {
  const avatar = profile.avatar_url
    ? `<img class="profile-avatar" src="${escapeAttr(profile.avatar_url)}" alt="" />`
    : `<div class="profile-avatar placeholder">${escapeHtml((profile.display_name || "?")[0].toUpperCase())}</div>`;

  const collegeBadge = profile.college_name
    ? `<span class="college-badge${profile.college_status === "pending" ? " pending" : ""}">${escapeHtml(profile.college_name)}${profile.college_status === "pending" ? " · pending" : ""}</span>`
    : "";

  const showUsername = alwaysShowUsername || !profile.is_anonymous;

  return `
    <div class="profile-header">
      ${avatar}
      <div>
        <div class="profile-name">${escapeHtml(profile.display_name)}</div>
        <div class="profile-meta-row">
          ${showUsername ? `<span class="profile-username">/u/${escapeHtml(profile.username)}</span>` : ""}
          ${collegeBadge}
        </div>
      </div>
    </div>
  `;
}

function renderStatsGrid(summary) {
  return `
    <div class="profile-stats-grid">
      <div class="stat-tile">
        <div class="num">${summary.completed_patterns}/${summary.total_patterns}</div>
        <div class="label">Patterns complete</div>
      </div>
      <div class="stat-tile">
        <div class="num">${summary.completed_steps}/${summary.total_steps}</div>
        <div class="label">Steps checked</div>
      </div>
    </div>
  `;
}

/* ── Own (editable) view ──────────────────────────────────────────── */

async function renderEditableProfile(me) {
  if (me.account_status !== "active") {
    contentEl.innerHTML = `<p class="state-message">Your account isn't fully active yet — visit the home page for details.</p>`;
    return;
  }

  let profile;
  try {
    profile = await api.profile.getMine();
  } catch (err) {
    contentEl.innerHTML = `<p class="state-message">Couldn't load your profile: ${escapeHtml(err.message)}</p>`;
    return;
  }

  contentEl.innerHTML = `
    ${renderProfileHeader(profile, true)}
    ${renderStatsGrid(profile.summary)}

    <div class="profile-section">
      <h2>Edit profile</h2>

      <div class="form-error" id="form-error"></div>

      <div class="field">
        <label for="username-input">Username</label>
        <input type="text" id="username-input" value="${escapeAttr(profile.username)}" />
      </div>

      <div class="field">
        <label for="leetcode-input">LeetCode profile (optional)</label>
        <input type="text" id="leetcode-input" value="${escapeAttr(profile.leetcode_url || "")}" placeholder="https://leetcode.com/u/yourname" />
      </div>

      <div class="toggle-row">
        <div>
          <div class="toggle-label">Anonymous mode</div>
          <div class="toggle-sublabel">Hide your name, avatar, college, and LeetCode link from your public profile and the leaderboard. Stats stay visible.</div>
        </div>
        <div class="toggle-switch${profile.is_anonymous ? " on" : ""}" id="anon-toggle" role="switch" aria-checked="${profile.is_anonymous}">
          <div class="knob"></div>
        </div>
      </div>

      <div class="save-bar">
        <button class="btn btn-primary" id="save-profile-btn">Save changes</button>
        <span class="save-indicator" id="profile-save-indicator">Saved</span>
      </div>
    </div>
  `;

  let isAnonymous = profile.is_anonymous;
  const anonToggle = document.getElementById("anon-toggle");
  anonToggle.addEventListener("click", () => {
    isAnonymous = !isAnonymous;
    anonToggle.classList.toggle("on", isAnonymous);
    anonToggle.setAttribute("aria-checked", String(isAnonymous));
  });

  const errorBox = document.getElementById("form-error");
  const saveBtn = document.getElementById("save-profile-btn");
  const indicator = document.getElementById("profile-save-indicator");

  saveBtn.addEventListener("click", async () => {
    errorBox.classList.remove("visible");

    const newUsername = document.getElementById("username-input").value.trim().toLowerCase();
    const newLeetcode = document.getElementById("leetcode-input").value.trim();

    saveBtn.disabled = true;
    try {
      const updated = await api.profile.update({
        username: newUsername,
        leetcode_url: newLeetcode,
        is_anonymous: isAnonymous,
      });

      indicator.textContent = "Saved";
      indicator.classList.add("visible");
      setTimeout(() => indicator.classList.remove("visible"), 1500);

      // Username may have changed — the URL needs to follow it.
      if (updated.username !== requestedUsername) {
        window.location.href = `/u/${encodeURIComponent(updated.username)}`;
      }
    } catch (err) {
      errorBox.textContent = err.message;
      errorBox.classList.add("visible");
    } finally {
      saveBtn.disabled = false;
    }
  });
}

function escapeAttr(str) {
  return escapeHtml(str).replace(/"/g, "&quot;");
}
