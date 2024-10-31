pg-run:
	sudo docker compose -f ./utils/driver/postgresql/docker-compose-pg.yml up -d
pg-start:
	sudo docker container start pg_auth_system_container
pg-exec:
	sudo docker exec -it pg_auth_system_container psql -U dwiw auth_system_go
pg-stop:
	sudo docker container stop pg_auth_system_container

rd-run:
	sudo docker run --name rd_auth_system -p 6379:6379 -d redis
rd-start:
	sudo docker container start rd_auth_system
rd-exec:
	sudo docker exec -it rd_auth_system redis-cli -h localhost
rd-stop:
	sudo docker container stop rd_auth_system

pg-exec-table-all:
	sudo docker exec -i pg_auth_system_container psql -U dwiw auth_system_go < internal/utils/postgres/all_tables.sql
pg-exec-table-users:
	sudo docker exec -i pg_auth_system_container psql -U dwiw auth_system_go < internal/utils/postgres/users.sql
pg-exec-table-email:
	sudo docker exec -i pg_auth_system_container psql -U dwiw auth_system_go < internal/utils/postgres/emails.sql
pg-exec-table-drop:
	sudo docker exec -i pg_auth_system_container psql -U dwiw auth_system_go < internal/utils/postgres/drop_all_tables.sql
pg-exec-table-privKey:
	sudo docker exec -i pg_auth_system_container psql -U dwiw auth_system_go < internal/utils/postgres/priv_key.sql
pg-exec-insert-marital-status:
	sudo docker exec -i pg_auth_system_container psql -U dwiw auth_system_go < internal/utils/postgres/insert_marital_status.sql

migrate-create:
	migrate create -ext sql -dir internal/database/migrations -seq init
migrate-up:
	migrate -path internal/database/migrations -database "postgresql://dwiw:secret@localhost:5432/auth_system_go?sslmode=disable" -verbose up $(v)
migrate-down:
	migrate -path internal/database/migrations -database "postgresql://dwiw:secret@localhost:5432/auth_system_go?sslmode=disable" -verbose down $(v)
migrate-force:
	migrate -path internal/database/migrations -database "postgresql://dwiw:secret@localhost:5432/auth_system_go?sslmode=disable" force $(v)

print:
	echo "Hi $(name)"
