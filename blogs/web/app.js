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
  // The backend renders the post server-side and returns it as raw HTML
  // (not JSON) — title and images are already embedded in the markup.
  return res.text();
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
  // Same as fetchBlogBySlug: raw HTML, not JSON.
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

// Renders a single blog post's HTML into `container` (a
// <article class="post__content">). `html` is the backend's fully
// server-rendered post markup (title heading and image URLs already
// embedded), so it's injected as-is rather than being assembled from parts.
function renderBlogPost(container, html) {
  container.innerHTML = "";

  const body = document.createElement("div");
  body.className = "post__body";
  // Trusted: this is the site owner's own backend-rendered content, not
  // user-submitted input, so no sanitization step is added here.
  body.innerHTML = html;

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
