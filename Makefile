.PHONY: *
DOCKER_NAME = dev-postgres

dockerup:
	limactl start docker

dockerdown: 
	limactl stop docker

postgresup:
	docker start ${DOCKER_NAME} \
	|| docker run --name ${DOCKER_NAME} \
	-e POSTGRES_PASSWORD=postgres \
	-e POSTGRES_DB=simple_bank \
	-p 5432:5432 -d postgres

postgresdown:
	docker stop ${DOCKER_NAME}
	docker rm ${DOCKER_NAME}

# createdb:
# 	docker exec -it ${DOCKER_NAME} createdb -U postgres -O postgres simple_bank

# dropdb:
# 	docker exec -it ${DOCKER_NAME} dropdb -U postgres simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" -verbose up

migrateupone:
	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" -verbose down

migratedownone:
	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" -verbose down 1
# migrateforce:
# 	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" force 1

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

runtest: postgresup createdb migrateup
	go test -v -cover ./...

clean: postgresdown

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/dubass83/simplebank/db/sqlc Store

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