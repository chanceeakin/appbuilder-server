
migrate-up:
	source .env && migrate -path db/migrations -database $$DATABASE_URL up;

migrate-down:
	source .env && migrate -path db/migrations -database $$DATABASE_URL down;

db-up:
	docker run -p 5432:5432 postgres

create-db:
	psql -h localhost -p 5432 -U postgres -W -c "create database appbuilder;"

run:
	go run .

build:
	go build -o appbuilder-server .