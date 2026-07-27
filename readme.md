Portfolio site projects written in HTML, CSS and bit of JS

## Running with Docker

### Build the image

```bash
docker build -t insanelyharsh/web-portfolio:latest .
```

### Run locally with Docker Compose

```bash
docker compose up -d
```

The site will be available at http://localhost:3000

### Push the image to Docker Hub

```bash
docker login
docker push insanelyharsh/web-portfolio:latest
```

### Build for multiple architectures (amd64 + arm64)

Requires Docker Buildx (bundled with Docker Desktop / recent Docker Engine).
Multi-arch images can't be loaded locally — they build straight to the registry:

```bash
docker buildx create --name web-portfolio-builder --use
docker buildx build --platform linux/amd64,linux/arm64 -t insanelyharsh/web-portfolio:latest --push .
```

Or simply:

```bash
make buildx
```

### Pull and run the image

```bash
docker pull insanelyharsh/web-portfolio:latest
docker run -d -p 3000:3000 insanelyharsh/web-portfolio:latest
```
