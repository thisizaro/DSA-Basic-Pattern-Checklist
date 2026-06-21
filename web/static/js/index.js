/**
 * index.js — powers the landing page (index.html). Decides between three
 * states: signed-out login screen, pending-review status screen, or
 * blocked status screen. Active users are redirected straight to the
 * checklist.
 */

const loginCard = document.getElementById("login-card");
const statusCard = document.getElementById("status-card");
const errorBox = document.getElementById("form-error");
const googleBtn = document.getElementById("google-signin-btn");
const statusLogoutBtn = document.getElementById("status-logout-btn");

googleBtn.href = api.auth.googleLoginUrl();

statusLogoutBtn.addEventListener("click", async () => {
  await api.auth.logout();
  window.location.reload();
});

init();

async function init() {
  showOAuthErrorIfPresent();

  let user;
  try {
    user = await api.auth.me();
  } catch {
    // Not signed in — show the normal login card (already visible by default).
    return;
  }

  if (user.account_status === "active") {
    window.location.href = "/checklist.html";
    return;
  }

  showStatusScreen(user.account_status);
}

function showOAuthErrorIfPresent() {
  const params = new URLSearchParams(window.location.search);
  const reason = params.get("auth_error");
  if (!reason) return;

  const messages = {
    invalid_state: "Login session expired or was tampered with. Please try again.",
    missing_code: "Google didn't return a login code. Please try again.",
    email_not_verified: "Your Google email isn't verified. Verify it with Google, then try again.",
    login_failed: "Something went wrong signing you in. Please try again.",
  };

  errorBox.textContent = messages[reason] || "Sign-in failed. Please try again.";
  errorBox.classList.add("visible");

  // Clean the error out of the URL so a refresh doesn't re-show it.
  params.delete("auth_error");
  const cleanUrl = window.location.pathname + (params.toString() ? `?${params}` : "");
  window.history.replaceState({}, "", cleanUrl);
}

function showStatusScreen(status) {
  loginCard.style.display = "none";
  statusCard.style.display = "block";

  const badge = document.getElementById("status-badge");
  const title = document.getElementById("status-title");
  const body = document.getElementById("status-body");
  const collegeForm = document.getElementById("college-name-form");

  if (status === "pending_review") {
    badge.textContent = "Pending review";
    title.textContent = "Your college is under review";
    body.innerHTML = `
      Looks like you're the first person from your college to sign up here —
      nice. Your account works, but your college badge is waiting on manual
      review before it's confirmed. If you haven't already, tell us your
      college's name below so the admin knows what to approve.
    `;
    collegeForm.style.display = "block";
    setupCollegeNameForm();
  } else {
    badge.textContent = "Access blocked";
    title.textContent = "College account required";
    body.innerHTML = `
      You're not allowed to log in with a non-college domain. If you're a
      developer who needs access for testing, contact the admin to get your
      account whitelisted — then sign in again.
    `;
  }
}

function setupCollegeNameForm() {
  const input = document.getElementById("college-name-input");
  const submitBtn = document.getElementById("college-name-submit");
  const errorBox = document.getElementById("college-name-error");

  submitBtn.addEventListener("click", async () => {
    errorBox.classList.remove("visible");
    const name = input.value.trim();

    submitBtn.disabled = true;
    try {
      await api.auth.nameCollege(name);
      submitBtn.textContent = "Submitted — thanks!";
      input.disabled = true;
    } catch (err) {
      errorBox.textContent = err.message;
      errorBox.classList.add("visible");
      submitBtn.disabled = false;
    }
  });
}
