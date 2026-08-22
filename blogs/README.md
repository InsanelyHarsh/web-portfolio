# Blogs

This is my blog site, split into two small pieces: a **backend** that stores and serves posts, and a **web** frontend that shows them.

## Tech stack

**Backend** (`blogs/backend`)
- Go, using the standard library's `net/http` (no framework)
- PostgreSQL as the database, accessed via `pgx` (no ORM)
- Posts are stored as raw Markdown and rendered to HTML on request (`gomarkdown` + `bluemonday` for sanitizing)
- Images/media are stored in Cloudflare R2 and proxied through the backend

**Web** (`blogs/web`)
- Plain HTML, CSS, and a bit of vanilla JS — no framework or build step
- A tiny Node server just serves the static files

## How it works

Blog posts live in a Postgres table as Markdown text. There's no admin UI yet, so posts are added straight into the DB — they were previously written on [Hashnode](https://hashnode.com/@insanelyharsh) and imported over from there.

When the frontend asks for a post, the backend pulls the Markdown from Postgres, converts it to sanitized HTML, and sends it back — as JSON for the blog list, or as HTML for a single post page. The frontend just fetches from the backend's endpoints and drops the result onto the page. Any images in a post are served through the backend too, which fetches them from Cloudflare R2.

Everything is public/read-only right now — no login or auth required to view posts.

## Not there yet

- No admin UI for writing/managing posts
- No comments
- No view counting

## Running locally

Both `blogs/backend` and `blogs/web` have their own `docker-compose.yml` and `Makefile`. From either folder:

```
make up      # start it
make logs    # tail logs
make down    # stop it
```

The backend also spins up its own Postgres container and applies DB migrations automatically on startup.

## Roadmap

Some things planned for down the line:

- SEO
- Blocking bots from crawling the site
- Rate limiting
- Database replication/backups

## Deployment

It's deployed using Docker — the same images built locally (`make build`/`make buildx`, which is why they're built for `linux/arm64` too) get pushed to Docker Hub and run as containers in production.

Currently everything runs on a Raspberry Pi 4 at home, sitting behind Nginx, and exposed to the internet through a Cloudflare Tunnel.
