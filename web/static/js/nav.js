/**
 * nav.js — renders the shared site header (brand + nav links + user menu)
 * into a `<header id="site-header">` placeholder. Used by every page that
 * has the full app chrome (checklist, profile, leaderboard).
 *
 * `user` is null for logged-out visitors (leaderboard is public) — in that
 * case the nav shows a "Sign in" link instead of the user tag/logout button.
 */
function renderNav(activePage, user) {
  const header = document.getElementById("site-header");
  if (!header) return;

  const links = [
    { href: "/checklist.html", label: "Checklist", key: "checklist" },
    { href: "/leaderboard.html", label: "Leaderboard", key: "leaderboard" },
  ];
  if (user) {
    links.push({ href: `/u/${encodeURIComponent(user.username)}`, label: "Profile", key: "profile" });
  }

  const linksHtml = links
    .map(
      (l) =>
        `<a class="nav-link${l.key === activePage ? " active" : ""}" href="${l.href}">${l.label}</a>`
    )
    .join("");

  const actionsHtml = user
    ? `
      <span class="user-tag">${escapeHtml(user.display_name)}</span>
      <button class="btn btn-ghost btn-sm" id="nav-logout-btn">Log out</button>
    `
    : `<a class="btn btn-ghost btn-sm" href="/index.html">Sign in</a>`;

  header.innerHTML = `
    <div class="shell">
      <a class="brand" href="/checklist.html" style="text-decoration:none;"><span class="dot"></span>DSA Tracker</a>
      <div class="header-actions">
        <nav class="nav-links">${linksHtml}</nav>
        ${actionsHtml}
      </div>
    </div>
  `;

  const logoutBtn = document.getElementById("nav-logout-btn");
  if (logoutBtn) {
    logoutBtn.addEventListener("click", async () => {
      await api.auth.logout();
      window.location.href = "/index.html";
    });
  }
}

function escapeHtml(str) {
  const div = document.createElement("div");
  div.textContent = str == null ? "" : str;
  return div.innerHTML;
}
