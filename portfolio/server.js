const http = require('http');
const fs = require('fs');
const path = require('path');

const ROOT = __dirname;

// Minimal, dependency-free .env loader: reads KEY=VALUE lines from a local
// .env file and fills in process.env for any key not already set (real env
// vars, e.g. from docker-compose, always win).
function loadEnv() {
  const envPath = path.join(ROOT, '.env');
  if (!fs.existsSync(envPath)) return;

  const lines = fs.readFileSync(envPath, 'utf8').split('\n');
  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith('#')) continue;

    const eq = trimmed.indexOf('=');
    if (eq === -1) continue;

    const key = trimmed.slice(0, eq).trim();
    let value = trimmed.slice(eq + 1).trim();
    if (
      (value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))
    ) {
      value = value.slice(1, -1);
    }

    if (key && !(key in process.env)) {
      process.env[key] = value;
    }
  }
}

loadEnv();

const PORT = process.env.PORT || 3000;

// Resume link is read from env so it can point at wherever the PDF is
// actually hosted (object storage in production) without touching index.html.
// Defaults to the file shipped alongside this server for local dev.
// Escaped for safe embedding in an HTML attribute: object storage URLs
// (e.g. presigned R2 links) commonly carry `&` in their query string.
const RESUME_URL = (process.env.RESUME_URL || './Harsh M Yadav SDE - Resume.pdf')
  .trim()
  .replaceAll('&', '&amp;')
  .replaceAll('"', '&quot;');
const INDEX_HTML_PATH = path.join(ROOT, 'index.html');

const CONTENT_TYPES = {
  '.html': 'text/html; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.js': 'application/javascript; charset=utf-8',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.pdf': 'application/pdf',
};

const server = http.createServer((req, res) => {
  const urlPath = req.url.split('?')[0];
  const requestPath = urlPath === '/' ? '/index.html' : urlPath;
  const safePath = path.normalize(decodeURIComponent(requestPath)).replace(/^(\.\.[/\\])+/, '');
  const filePath = path.join(ROOT, safePath);

  // Prevent path traversal outside the project root.
  if (!filePath.startsWith(ROOT)) {
    res.writeHead(403, { 'Content-Type': 'text/plain' });
    res.end('403 Forbidden');
    return;
  }

  fs.readFile(filePath, (err, data) => {
    if (err) {
      if (err.code === 'ENOENT') {
        res.writeHead(404, { 'Content-Type': 'text/plain' });
        res.end('404 Not Found');
      } else {
        res.writeHead(500, { 'Content-Type': 'text/plain' });
        res.end('500 Internal Server Error');
      }
      return;
    }

    const ext = path.extname(filePath).toLowerCase();
    const contentType = CONTENT_TYPES[ext] || 'application/octet-stream';

    // index.html ships with a __RESUME_URL__ placeholder so the resume link
    // can be configured via env rather than hardcoded into the static asset.
    const body = filePath === INDEX_HTML_PATH
      ? data.toString('utf8').replaceAll('__RESUME_URL__', RESUME_URL)
      : data;

    // No Cache-Control here previously meant Cloudflare fell back to its own
    // default edge-cache TTL for static extensions (hours), so a deploy could
    // update the origin while the CDN kept serving a stale cached copy.
    // `no-cache` forces a revalidation every time instead.
    res.writeHead(200, { 'Content-Type': contentType, 'Cache-Control': 'no-cache' });
    res.end(body);
  });
});

server.listen(PORT, () => {
  console.log(`Server running at http://localhost:${PORT}/`);
});
