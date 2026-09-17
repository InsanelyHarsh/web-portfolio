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

  enhanceCodeBlocks(body);
  enhanceHeadingAnchors(body);
  wrapTables(body);

  const heading = body.querySelector("h1, h2");
  document.title = heading ? `${heading.textContent} — Insane Blogs` : "Insane Blogs";

  const backLink = el("a", "post__back", "← Back to all posts");
  backLink.href = "blog_list.html";
  const backPara = el("p");
  backPara.appendChild(backLink);

  container.append(body, backPara);
}

const LANGUAGE_LABELS = {
  js: "JavaScript", javascript: "JavaScript",
  ts: "TypeScript", typescript: "TypeScript",
  py: "Python", python: "Python",
  go: "Go", golang: "Go",
  rb: "Ruby", ruby: "Ruby",
  rs: "Rust", rust: "Rust",
  java: "Java",
  c: "C", cpp: "C++", "c++": "C++",
  cs: "C#", csharp: "C#",
  php: "PHP",
  html: "HTML", xml: "XML",
  css: "CSS", scss: "SCSS",
  json: "JSON",
  yaml: "YAML", yml: "YAML",
  sh: "Shell", bash: "Bash", shell: "Shell", zsh: "Zsh",
  sql: "SQL",
  md: "Markdown", markdown: "Markdown",
  dockerfile: "Dockerfile",
  makefile: "Makefile",
  diff: "Diff",
  graphql: "GraphQL",
  kotlin: "Kotlin",
  swift: "Swift",
  plaintext: "Text", text: "Text", txt: "Text",
};

function titleCaseLangId(id) {
  return id.replace(/[-_]/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

function languageLabel(langId) {
  if (!langId) return "Text";
  return LANGUAGE_LABELS[langId.toLowerCase()] || titleCaseLangId(langId);
}

function copyCodeToClipboard(codeEl, button) {
  const resetLabel = "Copy";
  navigator.clipboard.writeText(codeEl.textContent).then(
    () => {
      button.textContent = "Copied";
      button.classList.add("code-block__copy--copied");
      clearTimeout(button._copyResetTimer);
      button._copyResetTimer = setTimeout(() => {
        button.textContent = resetLabel;
        button.classList.remove("code-block__copy--copied");
      }, 1500);
    },
    () => {
      button.textContent = "Failed";
      clearTimeout(button._copyResetTimer);
      button._copyResetTimer = setTimeout(() => {
        button.textContent = resetLabel;
      }, 1500);
    }
  );
}

function enhanceCodeBlocks(root) {
  root.querySelectorAll("pre > code").forEach((code) => {
    const pre = code.parentElement;
    if (pre.dataset.codeBlockEnhanced === "true") return;

    const langClass = [...code.classList].find((c) => c.startsWith("language-"));
    const langId = langClass ? langClass.slice("language-".length) : null;

    if (langId && window.hljs && !code.classList.contains("hljs")) {
      try {
        hljs.highlightElement(code);
      } catch (err) {
        console.warn("Syntax highlighting failed:", err);
      }
    }

    const header = el("div", "code-block__header");
    const copyButton = el("button", "code-block__copy", "Copy");
    copyButton.type = "button";
    copyButton.setAttribute("aria-label", "Copy code to clipboard");
    copyButton.addEventListener("click", () => copyCodeToClipboard(code, copyButton));

    header.append(el("span", "code-block__lang", languageLabel(langId)), copyButton);

    const wrapper = el("div", "code-block");
    pre.before(wrapper);
    wrapper.append(header, pre);
    pre.dataset.codeBlockEnhanced = "true";
  });
}

function enhanceHeadingAnchors(root) {
  root.querySelectorAll(":is(h1, h2, h3, h4, h5, h6)[id]").forEach((heading) => {
    if (heading.querySelector(":scope > .heading-anchor")) return;
    const anchor = el("a", "heading-anchor", "#");
    anchor.href = `#${heading.id}`;
    anchor.setAttribute("aria-label", "Link to this section");
    heading.appendChild(anchor);
  });
}

function wrapTables(root) {
  root.querySelectorAll("table").forEach((table) => {
    if (table.parentElement.classList.contains("post__table-wrapper")) return;
    const wrapper = el("div", "post__table-wrapper");
    table.before(wrapper);
    wrapper.appendChild(table);
  });
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
