ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: start stop clean run

start:
	./scripts/startElasticLlama.sh
run:
	go run main.go
stop:
	docker compose -f ./docker/docker-compose.yml down -v
	docker compose -f ./docker/docker-compose-elasticollama.yml down -v
clean:
	docker compose -f ./docker/docker-compose.yml down -v
	docker compose -f ./docker/docker-compose-elasticollama.yml down -v
	rm -rf ./docker/ollama ./docker/elasticsearch_data