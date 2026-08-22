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

const PORT = process.env.PORT || 8081;

// Guards against a malformed API_BASE (missing scheme, trailing slash) silently
// producing a relative URL in the browser instead of an absolute one.
function normalizeApiBase(value) {
  let v = value.trim().replace(/\/+$/, '');
  if (v && !/^https?:\/\//i.test(v)) {
    v = 'https://' + v;
  }
  return v;
}

const API_BASE = normalizeApiBase(process.env.API_BASE || 'http://localhost:8080');
const APP_JS_PATH = path.join(ROOT, 'app.js');

// Logs one line per request: timestamp, method, url, status, and (for
// failures) a short reason so issues are visible without re-running curl.
function logRequest(req, statusCode, detail) {
  const ts = new Date().toISOString();
  const base = `[${ts}] ${req.method} ${req.url} -> ${statusCode}`;
  console.log(detail ? `${base} (${detail})` : base);
}

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

// Writes a response, logs it, and (for failures) attaches the reason so
// issues are visible without re-running curl.
function respond(req, res, statusCode, contentType, body, detail) {
  res.writeHead(statusCode, { 'Content-Type': contentType });
  res.end(body);
  logRequest(req, statusCode, detail);
}

const server = http.createServer((req, res) => {
  const urlPath = req.url.split('?')[0];
  const requestPath = urlPath === '/' ? '/index.html' : urlPath;
  const safePath = path.normalize(decodeURIComponent(requestPath)).replace(/^(\.\.[/\\])+/, '');
  const filePath = path.join(ROOT, safePath);

  // Prevent path traversal outside the project root.
  if (!filePath.startsWith(ROOT)) {
    respond(req, res, 403, 'text/plain', '403 Forbidden', 'path traversal blocked');
    return;
  }

  fs.readFile(filePath, (err, data) => {
    if (err) {
      if (err.code === 'ENOENT') {
        respond(req, res, 404, 'text/plain', '404 Not Found', 'not found');
      } else {
        console.error(err.stack || err);
        respond(req, res, 500, 'text/plain', '500 Internal Server Error', err.message);
      }
      return;
    }

    const ext = path.extname(filePath).toLowerCase();
    const contentType = CONTENT_TYPES[ext] || 'application/octet-stream';

    // app.js ships with an __API_BASE__ placeholder so the backend URL can
    // be configured via env rather than hardcoded into the static asset.
    const body = filePath === APP_JS_PATH
      ? data.toString('utf8').replace('__API_BASE__', API_BASE)
      : data;

    respond(req, res, 200, contentType, body);
  });
});

server.listen(PORT, () => {
  console.log(`Server running at http://localhost:${PORT}/`);
});
