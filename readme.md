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

### Pull and run the image

```bash
docker pull insanelyharsh/web-portfolio:latest
docker run -d -p 3000:3000 insanelyharsh/web-portfolio:latest
```
