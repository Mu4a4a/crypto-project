.SILENT:

build:
	docker build -t crypto-project .

run: build
	docker compose up

migration-up:
	migrate -database "postgres://postgres:qwerty@172.19.0.1:5432/postgres?SSLMode=disable" -path ./migrations up