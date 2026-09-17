const http = require('http');
const fs = require('fs');
const path = require('path');

const ROOT = __dirname;

function loadEnv() {
  const envPath = path.join(ROOT, '.env');
  if (!fs.existsSync(envPath)) return;

  for (const line of fs.readFileSync(envPath, 'utf8').split('\n')) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith('#')) continue;

    const eq = trimmed.indexOf('=');
    if (eq === -1) continue;

    const key = trimmed.slice(0, eq).trim();
    let value = trimmed.slice(eq + 1).trim();
    if (/^"[^"]*"$|^'[^']*'$/.test(value)) value = value.slice(1, -1);

    if (key && !(key in process.env)) process.env[key] = value;
  }
}

loadEnv();

const PORT = process.env.PORT || 8081;

function normalizeApiBase(value) {
  let v = value.trim().replace(/\/+$/, '');
  if (v && !/^https?:\/\//i.test(v)) v = `https://${v}`;
  return v;
}

const API_BASE = normalizeApiBase(process.env.API_BASE || 'http://localhost:8080');
const APP_JS_PATH = path.join(ROOT, 'app.js');

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

function respond(req, res, statusCode, contentType, body) {
  res.writeHead(statusCode, { 'Content-Type': contentType });
  res.end(body);
  console.log(`${req.method} ${req.url} -> ${statusCode}`);
}

const server = http.createServer((req, res) => {
  const urlPath = req.url.split('?')[0];
  const requestPath = urlPath === '/' ? '/index.html' : urlPath;
  const safePath = path.normalize(decodeURIComponent(requestPath)).replace(/^(\.\.[/\\])+/, '');
  const filePath = path.join(ROOT, safePath);

  if (!filePath.startsWith(ROOT)) {
    respond(req, res, 403, 'text/plain', '403 Forbidden');
    return;
  }

  fs.readFile(filePath, (err, data) => {
    if (err) {
      const statusCode = err.code === 'ENOENT' ? 404 : 500;
      const body = statusCode === 404 ? '404 Not Found' : '500 Internal Server Error';
      respond(req, res, statusCode, 'text/plain', body);
      return;
    }

    const ext = path.extname(filePath).toLowerCase();
    const contentType = CONTENT_TYPES[ext] || 'application/octet-stream';
    const body = filePath === APP_JS_PATH ? data.toString('utf8').replace('__API_BASE__', API_BASE) : data;

    res.writeHead(200, { 'Content-Type': contentType, 'Cache-Control': 'no-cache' });
    res.end(body);
    console.log(`${req.method} ${req.url} -> 200`);
  });
});

server.listen(PORT, () => console.log(`Server running at http://localhost:${PORT}/`));
