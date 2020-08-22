
migrate-up:
	source .env && migrate -path db/migrations -database $$DATABASE_URL up;

migrate-down:
	source .env && migrate -path db/migrations -database $$DATABASE_URL down;

run:
	go run main.go