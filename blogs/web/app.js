const API_BASE = "__API_BASE__";

async function fetchBlogList() {
  const url = `${API_BASE}/blogs`;
  console.log(`[blogs] GET ${url}`);
  let res;
  try {
    res = await fetch(url);
  } catch (err) {
    console.error(`[blogs] network error: GET ${url}`, err);
    throw new Error(`Failed to fetch blog list: ${err.message}`);
  }
  if (!res.ok) {
    console.error(`[blogs] GET ${url} -> ${res.status} ${res.statusText}`);
    throw new Error(`Failed to fetch blog list: ${res.status} ${res.statusText}`);
  }
  return res.json();
}

async function fetchBlogBySlug(slug) {
  const url = `${API_BASE}/blogs/${encodeURIComponent(slug)}`;
  console.log(`[blogs] GET ${url}`);
  let res;
  try {
    res = await fetch(url);
  } catch (err) {
    console.error(`[blogs] network error: GET ${url}`, err);
    throw new Error(`Failed to fetch blog by slug "${slug}": ${err.message}`);
  }
  if (res.status === 404) {
    console.error(`[blogs] GET ${url} -> 404 (no blog for slug "${slug}")`);
    throw new Error(`Blog not found for slug "${slug}"`);
  }
  if (!res.ok) {
    console.error(`[blogs] GET ${url} -> ${res.status} ${res.statusText}`);
    throw new Error(`Failed to fetch blog by slug "${slug}": ${res.status} ${res.statusText}`);
  }
  return res.json();
}

async function fetchBlogById(id) {
  const url = `${API_BASE}/blogs/id/${encodeURIComponent(id)}`;
  console.log(`[blogs] GET ${url}`);
  let res;
  try {
    res = await fetch(url);
  } catch (err) {
    console.error(`[blogs] network error: GET ${url}`, err);
    throw new Error(`Failed to fetch blog by id "${id}": ${err.message}`);
  }
  if (res.status === 404) {
    console.error(`[blogs] GET ${url} -> 404 (no blog for id "${id}")`);
    throw new Error(`Blog not found for id "${id}"`);
  }
  if (!res.ok) {
    console.error(`[blogs] GET ${url} -> ${res.status} ${res.statusText}`);
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
  document.title = `${blog.title} — Insane Blogs`;

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
