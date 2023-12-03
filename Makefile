.PHONY: *
DOCKER_NAME = dev-postgres
postgresup:
	colima start
	docker stop ${DOCKER_NAME} || true
	docker rm ${DOCKER_NAME} || true
	docker run --name ${DOCKER_NAME} -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres
	sleep 5

postgresdown:
	docker stop ${DOCKER_NAME}
	docker rm ${DOCKER_NAME}
	colima stop

createdb:
	docker exec -it ${DOCKER_NAME} createdb -U postgres -O postgres simple_bank

dropdb:
	docker exec -it ${DOCKER_NAME} dropdb -U postgres simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" -verbose down

# migrateforce:
# 	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" force 1

sqlc:
	sqlc generate

test: postgresup createdb migrateup
	go test -v -cover ./...

clean: postgresdown