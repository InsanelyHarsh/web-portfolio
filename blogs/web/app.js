const API_BASE = "http://localhost:8080";

async function fetchBlogList() {
  const res = await fetch(`${API_BASE}/blogs`);
  if (!res.ok) {
    throw new Error(`Failed to fetch blog list: ${res.status} ${res.statusText}`);
  }
  return res.json();
}

async function fetchBlogBySlug(slug) {
  const res = await fetch(`${API_BASE}/blogs/${encodeURIComponent(slug)}`);
  if (res.status === 404) {
    throw new Error(`Blog not found for slug "${slug}"`);
  }
  if (!res.ok) {
    throw new Error(`Failed to fetch blog by slug "${slug}": ${res.status} ${res.statusText}`);
  }
  return res.json();
}

async function fetchBlogById(id) {
  const res = await fetch(`${API_BASE}/blogs/id/${encodeURIComponent(id)}`);
  if (res.status === 404) {
    throw new Error(`Blog not found for id "${id}"`);
  }
  if (!res.ok) {
    throw new Error(`Failed to fetch blog by id "${id}": ${res.status} ${res.statusText}`);
  }
  return res.json();
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

// Renders a single Blog into `container` (a <article class="post__content">).
function renderBlogPost(container, blog) {
  container.innerHTML = "";
  document.title = `${blog.title} — insanelyharsh blogs`;

  const title = document.createElement("h1");
  title.className = "post__title";
  title.textContent = blog.title;

  const meta = document.createElement("p");
  meta.className = "post__meta";
  meta.textContent = new Date(blog.createdAt).toLocaleDateString(undefined, {
    year: "numeric",
    month: "long",
    day: "numeric",
  });

  const body = document.createElement("div");
  body.className = "post__body";
  // Trusted: this is the site owner's own backend-rendered content, not
  // user-submitted input, so no sanitization step is added here.
  body.innerHTML = blog.content;

  const backPara = document.createElement("p");
  const backLink = document.createElement("a");
  backLink.href = "blog_list.html";
  backLink.className = "post__back";
  backLink.textContent = "← Back to all posts";
  backPara.appendChild(backLink);

  container.append(title, meta, body, backPara);
}

// Renders a fetch/validation error into `container`.
function renderError(container, err) {
  container.innerHTML = "";
  const p = document.createElement("p");
  p.className = "error-message";
  p.textContent = (err && err.message) || "Something went wrong.";
  container.appendChild(p);
}
