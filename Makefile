
build:
	go build .

install:
	go install .

docker:
	docker compose --env-file ./.env up -d

docker-recreate:
	docker compose --env-file ./.env up -d --force-recreate