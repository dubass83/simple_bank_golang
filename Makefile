.PHONY: *
ENV = dev
DB_URL = "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable"

docker_up:
	limactl start docker

docker_down: 
	limactl stop docker

postgres_up:
	docker start ${ENV}-postgres \
	|| docker run --name ${ENV}-postgres \
	-e POSTGRES_PASSWORD=postgres \
	-e POSTGRES_DB=simple_bank \
	-p 5432:5432 -d postgres

postgres_down:
	docker stop ${ENV}-postgres
	docker rm ${ENV}-postgres

redis_up:
	docker start ${ENV}-redis \
	|| docker run --name ${ENV}-redis \
	-p 6379:6379 -d redis:7.2-alpine

redis_down:
	docker stop ${ENV}-redis
	docker rm ${ENV}-redis

# createdb:
# 	docker exec -it ${ENV}-postgres createdb -U postgres -O postgres simple_bank

# dropdb:
# 	docker exec -it ${ENV}-postgres dropdb -U postgres simple_bank

migrate_up:
	migrate -path db/migration -database ${DB_URL} -verbose up

migrate_up1:
	migrate -path db/migration -database ${DB_URL} -verbose up 1

migrate_down:
	migrate -path db/migration -database ${DB_URL} -verbose down

migrate_down1:
	migrate -path db/migration -database ${DB_URL} -verbose down 1
# migrateforce:
# 	migrate -path db/migration -database ${DB_URL} force 1

sqlc:
	sqlc generate

test:
	go test -v -cover -short ./...

run_test: postgres_up createdb migrate_up
	go test -v -cover ./...

clean: postgres_down redis_down docker_down

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/dubass83/simplebank/db/sqlc Store
	mockgen -package mockwk -destination worker/mock/distributer.go github.com/dubass83/simplebank/worker TaskDistributor

build:
	docker build -t simple-bank -f Dockerfile .

proto:
	rm -f pb/*
	rm -f docs/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
      --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	  --grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	  --openapiv2_out=docs/swagger --openapiv2_opt=allow_merge=true,merge_file_name=simple_bank \
      proto/*.proto
	  statik -src=./docs/swagger -dest=./docs

evans:
	evans -p 9090 -r repl

db_docs:
	dbdocs build docs/db.dbml

db_schema:
	dbml2sql --postgres -o docs/schema.sql docs/db.dbml

new_migration:
	migrate create -ext sql -dir db/migration -seq ${name}