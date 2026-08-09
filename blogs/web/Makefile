IMAGE := insanelyharsh/blogs-web:latest
PLATFORMS := linux/amd64,linux/arm64
BUILDER := blogs-web

.PHONY: build push pull up down buildx logs restart ps

build:
	docker build -t $(IMAGE) .

push:
	docker push $(IMAGE)

# builds & pushes a multi-arch (amd64 + arm64) image in one go
buildx:
	docker buildx inspect $(BUILDER) >/dev/null 2>&1 || docker buildx create --name $(BUILDER) --use
	docker buildx use $(BUILDER)
	docker buildx build --platform $(PLATFORMS) -t $(IMAGE) --push .

pull:
	docker pull $(IMAGE)

up:
	docker compose up -d

down:
	docker compose down

restart:
	docker compose restart

logs:
	docker compose logs -f

ps:
	docker compose ps
