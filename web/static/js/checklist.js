/**
 * checklist.js — renders the pattern checklist as a compact, filterable
 * table. Each row shows name + topic badge + a 4-dot progress indicator;
 * clicking a row expands an inline detail panel with the full checklist,
 * the reference question link, and notes.
 */

const contentEl = document.getElementById("content");
const summaryBar = document.getElementById("summary-bar");
const statPatterns = document.getElementById("stat-patterns");
const statSteps = document.getElementById("stat-steps");
const progressFill = document.getElementById("progress-fill");

const filterTrigger = document.getElementById("filter-trigger");
const filterTriggerLabel = document.getElementById("filter-trigger-label");
const filterPanel = document.getElementById("filter-panel");
const filterOptionsEl = document.getElementById("filter-options");
const filterSelectAll = document.getElementById("filter-select-all");
const filterClear = document.getElementById("filter-clear");
const searchInput = document.getElementById("search-input");

const STEP_KEYS = ["understood", "explained", "solved_blind", "solved_after_gap"];
const STEP_LABELS = {
  understood: "Understood conceptually",
  explained: "Explained on paper",
  solved_blind: "Solved in Notepad cold",
  solved_after_gap: "Re-solved after 3 days",
};

// Full unfiltered data, kept in memory so filtering/search re-render
// without re-fetching.
let allTopics = [];
// Set of topic IDs currently checked in the filter — starts as "all".
let selectedTopicIds = new Set();
// Pattern ID currently expanded (only one open at a time keeps the table tidy).
let expandedPatternId = null;
// Debounce timers for notes textareas, keyed by pattern ID.
const notesSaveTimers = new Map();
const NOTES_DEBOUNCE_MS = 800;

init();

async function init() {
  let user;
  try {
    user = await api.auth.me();
  } catch {
    window.location.href = "/index.html";
    return;
  }

  if (user.account_status !== "active") {
    window.location.href = "/index.html";
    return;
  }

  renderNav("checklist", user);
  setupFilterDropdown();
  searchInput.addEventListener("input", renderTable);

  await loadChecklist();
}

async function loadChecklist() {
  try {
    const [topics, summary] = await Promise.all([
      api.checklist.getAll(),
      api.checklist.getSummary(),
    ]);
    allTopics = topics || [];
    selectedTopicIds = new Set(allTopics.map((t) => t.id));

    renderSummary(summary);
    renderFilterOptions();
    renderTable();
  } catch (err) {
    contentEl.innerHTML = `<p class="state-message">Couldn't load checklist: ${escapeHtml(err.message)}</p>`;
  }
}

function renderSummary(summary) {
  summaryBar.style.display = "flex";
  statPatterns.textContent = `${summary.completed_patterns}/${summary.total_patterns}`;
  statSteps.textContent = `${summary.completed_steps}/${summary.total_steps}`;
  const pct = summary.total_steps > 0
    ? Math.round((summary.completed_steps / summary.total_steps) * 100)
    : 0;
  progressFill.style.width = `${pct}%`;
}

/* ── Filter dropdown ──────────────────────────────────────────────── */

function setupFilterDropdown() {
  filterTrigger.addEventListener("click", () => {
    const isOpen = filterPanel.classList.toggle("open");
    filterTrigger.setAttribute("aria-expanded", String(isOpen));
  });

  document.addEventListener("click", (e) => {
    if (!filterPanel.contains(e.target) && !filterTrigger.contains(e.target)) {
      filterPanel.classList.remove("open");
      filterTrigger.setAttribute("aria-expanded", "false");
    }
  });

  filterSelectAll.addEventListener("click", () => {
    selectedTopicIds = new Set(allTopics.map((t) => t.id));
    renderFilterOptions();
    renderTable();
  });

  filterClear.addEventListener("click", () => {
    selectedTopicIds = new Set();
    renderFilterOptions();
    renderTable();
  });
}

function renderFilterOptions() {
  filterOptionsEl.innerHTML = "";

  allTopics.forEach((topic) => {
    const patterns = topic.patterns || [];
    const label = document.createElement("label");
    label.className = "filter-option";

    const checkbox = document.createElement("input");
    checkbox.type = "checkbox";
    checkbox.checked = selectedTopicIds.has(topic.id);
    checkbox.addEventListener("change", () => {
      if (checkbox.checked) {
        selectedTopicIds.add(topic.id);
      } else {
        selectedTopicIds.delete(topic.id);
      }
      updateFilterTriggerLabel();
      renderTable();
    });

    const text = document.createElement("span");
    text.textContent = topic.title;

    const count = document.createElement("span");
    count.className = "count";
    count.textContent = patterns.length;

    label.appendChild(checkbox);
    label.appendChild(text);
    label.appendChild(count);
    filterOptionsEl.appendChild(label);
  });

  updateFilterTriggerLabel();
}

function updateFilterTriggerLabel() {
  const total = allTopics.length;
  const selected = selectedTopicIds.size;
  if (selected === total) {
    filterTriggerLabel.textContent = "All topics";
  } else if (selected === 0) {
    filterTriggerLabel.textContent = "No topics selected";
  } else {
    filterTriggerLabel.textContent = `${selected} of ${total} topics`;
  }
}

/* ── Table rendering ──────────────────────────────────────────────── */

function isComplete(progress) {
  return Boolean(
    progress &&
      progress.understood &&
      progress.explained &&
      progress.solved_blind &&
      progress.solved_after_gap
  );
}

function stepsDone(progress) {
  if (!progress) return 0;
  return STEP_KEYS.reduce((n, key) => n + (progress[key] ? 1 : 0), 0);
}

function getFilteredRows() {
  const query = searchInput.value.trim().toLowerCase();
  const rows = [];

  allTopics.forEach((topic) => {
    if (!selectedTopicIds.has(topic.id)) return;

    (topic.patterns || []).forEach((pattern) => {
      if (query) {
        const haystack = `${pattern.name} ${pattern.question_title}`.toLowerCase();
        if (!haystack.includes(query)) return;
      }
      rows.push({ pattern, topicTitle: topic.title });
    });
  });

  return rows;
}

function renderTable() {
  const rows = getFilteredRows();

  if (allTopics.length === 0) {
    contentEl.innerHTML = `<p class="state-message">No topics found. Run the seed migration against your database.</p>`;
    return;
  }

  if (rows.length === 0) {
    contentEl.innerHTML = `<p class="state-message">No patterns match the current filter/search.</p>`;
    return;
  }

  const table = document.createElement("table");
  table.className = "pattern-table";
  table.innerHTML = `
    <thead>
      <tr>
        <th></th>
        <th>Pattern</th>
        <th class="col-topic">Topic</th>
        <th class="col-steps">Progress</th>
      </tr>
    </thead>
    <tbody id="pattern-tbody"></tbody>
  `;
  contentEl.innerHTML = "";
  contentEl.appendChild(table);

  const tbody = table.querySelector("#pattern-tbody");
  rows.forEach(({ pattern, topicTitle }) => {
    tbody.appendChild(buildPatternRow(pattern, topicTitle));
    tbody.appendChild(buildDetailRow(pattern));
  });
}

function buildPatternRow(pattern, topicTitle) {
  const tr = document.createElement("tr");
  tr.className = "pattern-row";
  tr.dataset.patternId = pattern.id;
  if (isComplete(pattern.progress)) tr.classList.add("complete");
  if (expandedPatternId === pattern.id) tr.classList.add("expanded");

  const done = stepsDone(pattern.progress);
  const dots = STEP_KEYS.map((key, i) => `<span class="${i < done ? "filled" : ""}"></span>`).join("");

  tr.innerHTML = `
    <td><span class="row-expand-icon">▸</span></td>
    <td class="row-name">${escapeHtml(pattern.name)}</td>
    <td class="col-topic"><span class="row-topic-badge">${escapeHtml(topicTitle)}</span></td>
    <td class="col-steps"><span class="row-steps-dots">${dots}</span></td>
  `;

  tr.addEventListener("click", () => {
    expandedPatternId = expandedPatternId === pattern.id ? null : pattern.id;
    renderTable();
  });

  return tr;
}

function buildDetailRow(pattern) {
  const tr = document.createElement("tr");
  tr.className = "pattern-detail-row";
  if (expandedPatternId === pattern.id) tr.classList.add("open");

  const progress = pattern.progress || {};

  const checklistHtml = STEP_KEYS.map(
    (key, i) => `
      <label class="check-item${progress[key] ? " checked" : ""}" data-key="${key}">
        <input type="checkbox" ${progress[key] ? "checked" : ""} />
        <span class="step-badge">${i + 1}</span>
        <span>${STEP_LABELS[key]}</span>
      </label>
    `
  ).join("");

  tr.innerHTML = `
    <td colspan="4">
      <div class="detail-idea">${escapeHtml(pattern.core_idea)}</div>
      <div class="detail-question">
        <a href="${escapeAttr(pattern.question_url)}" target="_blank" rel="noopener noreferrer">→ ${escapeHtml(pattern.question_title)}</a>
      </div>
      <div class="checklist-grid">${checklistHtml}</div>
      <button type="button" class="notes-toggle">${progress.notes ? "− Notes" : "+ Notes"}</button>
      <span class="save-indicator">Saved</span>
      <div class="notes-area${progress.notes ? " open" : ""}">
        <div class="field" style="margin-bottom: 0;">
          <textarea placeholder="Edge cases you missed, time complexity notes, anything to remember...">${escapeHtml(progress.notes || "")}</textarea>
        </div>
      </div>
    </td>
  `;

  // Stop row-toggle clicks from firing when interacting with controls
  // inside the expanded detail panel.
  tr.addEventListener("click", (e) => e.stopPropagation());

  const checkItems = tr.querySelectorAll(".check-item");
  checkItems.forEach((item) => {
    const checkbox = item.querySelector("input");
    checkbox.addEventListener("change", () => {
      item.classList.toggle("checked", checkbox.checked);
      saveProgress(pattern.id, tr);
    });
  });

  const notesToggle = tr.querySelector(".notes-toggle");
  const notesArea = tr.querySelector(".notes-area");
  notesToggle.addEventListener("click", () => {
    const isOpen = notesArea.classList.toggle("open");
    notesToggle.textContent = isOpen ? "− Notes" : "+ Notes";
  });

  const textarea = tr.querySelector("textarea");
  textarea.addEventListener("input", () => {
    clearTimeout(notesSaveTimers.get(pattern.id));
    const timer = setTimeout(() => saveProgress(pattern.id, tr), NOTES_DEBOUNCE_MS);
    notesSaveTimers.set(pattern.id, timer);
  });

  return tr;
}

async function saveProgress(patternId, detailRow) {
  const checkboxes = detailRow.querySelectorAll(".check-item input");
  const values = {};
  checkboxes.forEach((cb) => {
    const key = cb.closest(".check-item").dataset.key;
    values[key] = cb.checked;
  });
  values.notes = detailRow.querySelector("textarea").value;

  const indicator = detailRow.querySelector(".save-indicator");

  try {
    const updated = await api.checklist.updateProgress(patternId, values);

    // Update in-memory data so the steps-dots / complete state stay correct
    // if the row re-renders without a full reload.
    for (const topic of allTopics) {
      const p = (topic.patterns || []).find((p) => p.id === patternId);
      if (p) {
        p.progress = updated;
        break;
      }
    }

    flashSaved(indicator);
    refreshSummaryQuietly();
    updateRowVisuals(patternId);
  } catch (err) {
    indicator.textContent = "Save failed";
    indicator.classList.add("visible");
  }
}

// Updates just the steps-dots / complete class on the (collapsed) row for
// this pattern, without losing the open detail panel or scroll position.
function updateRowVisuals(patternId) {
  const row = document.querySelector(`.pattern-row[data-pattern-id="${cssEscape(patternId)}"]`);
  if (!row) return;

  let progress = null;
  for (const topic of allTopics) {
    const p = (topic.patterns || []).find((p) => p.id === patternId);
    if (p) {
      progress = p.progress;
      break;
    }
  }

  row.classList.toggle("complete", isComplete(progress));
  const done = stepsDone(progress);
  const dots = row.querySelectorAll(".row-steps-dots span");
  dots.forEach((dot, i) => dot.classList.toggle("filled", i < done));
}

function flashSaved(indicator) {
  indicator.textContent = "Saved";
  indicator.classList.add("visible");
  setTimeout(() => indicator.classList.remove("visible"), 1200);
}

async function refreshSummaryQuietly() {
  try {
    const summary = await api.checklist.getSummary();
    renderSummary(summary);
  } catch {
    // Non-critical — summary will catch up on next full page load.
  }
}

function escapeAttr(str) {
  return escapeHtml(str).replace(/"/g, "&quot;");
}

// Minimal CSS.escape fallback for building an attribute selector safely
// (pattern IDs are UUIDs so this is mostly defensive).
function cssEscape(str) {
  if (window.CSS && window.CSS.escape) return window.CSS.escape(str);
  return String(str).replace(/[^a-zA-Z0-9_-]/g, "\\$&");
}
