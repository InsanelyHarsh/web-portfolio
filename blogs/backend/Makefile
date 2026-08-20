IMAGE := insanelyharsh/blogs-backend:latest
PLATFORMS := linux/amd64,linux/arm64
BUILDER := blogs-backend

.PHONY: build push pull up down buildx logs restart ps postgres_connect

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

# opens a psql shell against the running postgres container, authenticated
# as POSTGRES_USER/POSTGRES_DB from .env (plain `psql` inside the container
# defaults to the OS user, which isn't a valid postgres role).
postgres_connect:
	set -a; . ./.env; set +a; docker compose exec postgres psql -U "$$POSTGRES_USER" -d "$$POSTGRES_DB"
