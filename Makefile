PLATFORM=$(shell uname -m)
DATETIME=$(shell date "+%Y%m%d%H%M%S")

msgsvr:
	go build -o bin/main src/main.go
	chmod +x bin/main
clean:
	$(RM) tmp/* $(TARGET) 

PHONY: run
run:
	go run ./src/main.go --config=./config/config-test.toml


PHONY: docker-run
docker-run:
	docker compose up -d

PHONY: docker-stop
docker stop:
	docker compose down