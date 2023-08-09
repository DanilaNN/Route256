
build-all:
	cd checkout && GOOS=linux GOARCH=amd64 make build
	cd loms && GOOS=linux GOARCH=amd64 make build
	cd notifications && GOOS=linux GOARCH=amd64 make build

run-all: build-all
	#sudo docker compose up --force-recreate --build
	#docker-compose up --force-recreate --build
	docker compose -f docker-compose-kafka.yml -f docker-compose.yml -f ./infra/metrics/docker-compose.yml -f ./infra/tracing/docker-compose.yml -f ./infra/logger/docker-compose.yml up --force-recreate --build

precommit:
	cd checkout && make precommit
	cd loms && make precommit
	cd notifications && make precommit
