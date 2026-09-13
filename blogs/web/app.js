// Guards against a misconfigured API_BASE env var (e.g. missing the "https://"
// scheme, or a trailing slash). Without this, a value like
// "blogs.insanelyharsh.com/api/" gets used as `${API_BASE}/blogs`, which
// fetch() then treats as a path *relative to the current page* instead of an
// absolute URL.
function normalizeApiBase(raw) {
  let base = (raw || "").trim();
  if (base && !/^https?:\/\//i.test(base) && !base.startsWith("/")) {
    base = `https://${base}`;
  }
  return base.replace(/\/+$/, "");
}

const API_BASE = normalizeApiBase("__API_BASE__");

// Shared GET helper: logs the request, distinguishes network failures from
// HTTP error statuses, and gives 404s a caller-supplied message.
async function fetchOrThrow(url, { label, notFoundMessage } = {}) {
  console.log(`[blogs] GET ${url}`);
  let res;
  try {
    res = await fetch(url);
  } catch (err) {
    console.error(`[blogs] network error: GET ${url}`, err);
    throw new Error(`Failed to fetch ${label}: ${err.message}`);
  }
  if (notFoundMessage && res.status === 404) {
    console.error(`[blogs] GET ${url} -> 404 (${notFoundMessage})`);
    throw new Error(notFoundMessage);
  }
  if (!res.ok) {
    console.error(`[blogs] GET ${url} -> ${res.status} ${res.statusText}`);
    throw new Error(`Failed to fetch ${label}: ${res.status} ${res.statusText}`);
  }
  return res;
}

async function fetchBlogList() {
  const res = await fetchOrThrow(`${API_BASE}/blogs`, { label: "blog list" });
  return res.json();
}

// The backend renders the post server-side and returns it as raw HTML
// (not JSON) — title and images are already embedded in the markup.
async function fetchBlogBySlug(slug) {
  const res = await fetchOrThrow(`${API_BASE}/blogs/${encodeURIComponent(slug)}`, {
    label: `blog by slug "${slug}"`,
    notFoundMessage: `Blog not found for slug "${slug}"`,
  });
  return res.text();
}

// Renders a list of BlogListItem into `container` as clickable cards.
function renderBlogList(container, items) {
  container.innerHTML = "";

  if (!items || items.length === 0) {
    const empty = document.createElement("p");
    empty.className = "error-message";
    empty.textContent = "No posts yet.";
    container.appendChild(empty);
    return;
  }

  items.forEach((item) => {
    const card = document.createElement("a");
    card.className = "blog-card";
    card.href = `blog_content.html?slug=${encodeURIComponent(item.slug)}`;

    const title = document.createElement("h2");
    title.className = "blog-card__title";
    title.textContent = item.title;

    const excerpt = document.createElement("p");
    excerpt.className = "blog-card__excerpt";
    excerpt.textContent = item.excerpt;

    card.append(title, excerpt);
    container.appendChild(card);
  });
}

// Renders a post's rendered HTML (from fetchBlogBySlug) into `container`
// (a <article class="post__content">). The backend no longer sends
// title/date separately — the content's own first heading (if any) doubles
// as the visual title, and is used for document.title too.
function renderBlogPost(container, html) {
  container.innerHTML = "";

  const body = document.createElement("div");
  body.className = "post__body";
  // Trusted: this is the site owner's own backend-rendered content, not
  // user-submitted input, so no sanitization step is added here.
  body.innerHTML = html;

  const heading = body.querySelector("h1, h2");
  document.title = heading ? `${heading.textContent} — Insane Blogs` : "Insane Blogs";

  const backPara = document.createElement("p");
  const backLink = document.createElement("a");
  backLink.href = "blog_list.html";
  backLink.className = "post__back";
  backLink.textContent = "← Back to all posts";
  backPara.appendChild(backLink);

  container.append(body, backPara);
}

// Renders a fetch/validation error into `container`.
function renderError(container, err) {
  container.innerHTML = "";
  const p = document.createElement("p");
  p.className = "error-message";
  p.textContent = (err && err.message) || "Something went wrong.";
  container.appendChild(p);
}

// Theme toggle (dark/light, persisted in localStorage).
// The actual theme is applied before first paint by an inline script in
// <head> on each page; this just wires up the button and keeps it in sync.
const THEME_KEY = "blog-theme";

function currentTheme() {
  return document.documentElement.getAttribute("data-theme") === "light" ? "light" : "dark";
}

function updateThemeToggleUI(theme) {
  const toggle = document.getElementById("themeToggle");
  if (!toggle) return;
  toggle.textContent = theme === "dark" ? "☀️" : "🌙";
  toggle.setAttribute("aria-label", theme === "dark" ? "Switch to light mode" : "Switch to dark mode");
}

function setTheme(theme) {
  document.documentElement.setAttribute("data-theme", theme);
  localStorage.setItem(THEME_KEY, theme);
  updateThemeToggleUI(theme);
}

function initThemeToggle() {
  updateThemeToggleUI(currentTheme());
  const toggle = document.getElementById("themeToggle");
  if (toggle) {
    toggle.addEventListener("click", () => setTheme(currentTheme() === "dark" ? "light" : "dark"));
  }
}

document.addEventListener("DOMContentLoaded", initThemeToggle);
