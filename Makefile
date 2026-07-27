IMAGE := insanelyharsh/web-portfolio:latest
PLATFORMS := linux/amd64,linux/arm64
BUILDER := web-portfolio-builder

.PHONY: build push pull up down buildx

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
