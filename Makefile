postgresup:
	docker run --name dev-postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres

postgresdown:
	docker stop dev-postgres
	docker rm dev-postgres

createdb:
	docker exec -it dev-postgres createdb -U postgres -O postgres simple_bank

dropdb:
	docker exec -it dev-postgres dropdb -U postgres simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" -verbose down

# migrateforce:
# 	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" force 1

sqlc:
	sqlc generate

.PHONY: createdb dropdb postgresup postgresdown migratedown migrateup sqlc