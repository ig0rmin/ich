compose:
	docker-compose up -d

build:
	go build -o server ./cmd

psql:
	PGPASSWORD=$(ICH_DB_PASSWORD) psql -h $(ICH_DB_HOST) -p $(ICH_DB_PORT) -U $(ICH_DB_USER) $(ICH_DB_NAME) 

.PHONY: compose psql
