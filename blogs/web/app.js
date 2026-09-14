function normalizeApiBase(raw) {
  let base = (raw || "").trim();
  if (base && !/^https?:\/\//i.test(base) && !base.startsWith("/")) {
    base = `https://${base}`;
  }
  return base.replace(/\/+$/, "");
}

const API_BASE = normalizeApiBase("__API_BASE__");

function el(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text !== undefined) node.textContent = text;
  return node;
}

async function fetchOrThrow(url, { label, notFoundMessage } = {}) {
  const res = await fetch(url).catch((err) => {
    throw new Error(`Failed to fetch ${label}: ${err.message}`);
  });
  if (notFoundMessage && res.status === 404) throw new Error(notFoundMessage);
  if (!res.ok) throw new Error(`Failed to fetch ${label}: ${res.status} ${res.statusText}`);
  return res;
}

async function fetchBlogList() {
  const res = await fetchOrThrow(`${API_BASE}/blogs`, { label: "blog list" });
  return res.json();
}

async function fetchBlogBySlug(slug) {
  const res = await fetchOrThrow(`${API_BASE}/blogs/${encodeURIComponent(slug)}`, {
    label: `blog by slug "${slug}"`,
    notFoundMessage: `Blog not found for slug "${slug}"`,
  });
  return res.text();
}

function renderBlogList(container, items) {
  container.innerHTML = "";

  if (!items || items.length === 0) {
    container.appendChild(el("p", "error-message", "No posts yet."));
    return;
  }

  items.forEach((item) => {
    const card = el("a", "blog-card");
    card.href = `blog_content.html?slug=${encodeURIComponent(item.slug)}`;
    card.append(el("h2", "blog-card__title", item.title), el("p", "blog-card__excerpt", item.excerpt));
    container.appendChild(card);
  });
}

function renderBlogPost(container, html) {
  container.innerHTML = "";

  const body = el("div", "post__body");
  body.innerHTML = html;

  const heading = body.querySelector("h1, h2");
  document.title = heading ? `${heading.textContent} — Insane Blogs` : "Insane Blogs";

  const backLink = el("a", "post__back", "← Back to all posts");
  backLink.href = "blog_list.html";
  const backPara = el("p");
  backPara.appendChild(backLink);

  container.append(body, backPara);
}

function renderError(container, err) {
  container.innerHTML = "";
  container.appendChild(el("p", "error-message", (err && err.message) || "Something went wrong."));
}

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
