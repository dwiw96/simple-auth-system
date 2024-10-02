pgRun:
	sudo docker compose -f ./utils/driver/postgresql/docker-compose-pg.yml up -d
pgStart:
	sudo docker container start pg_auth_system_container
pgExec:
	sudo docker exec -it pg_auth_system_container psql -U dwiw auth_system_go
pgStop:
	sudo docker container stop pg_auth_system_container

rdRun:
	sudo docker run --name rd_auth_system -p 6379:6379 -d redis
rdStart:
	sudo docker container start rd_auth_system
rdExec:
	sudo docker exec -it rd_auth_system redis-cli -h localhost
rdStop:
	sudo docker container stop rd_auth_system

pgMigrateAll:
	sudo docker exec -i pg_auth_system_container psql -U dwiw auth_system_go < internal/utils/postgres/all_tables.sql
pgMigrateUsers:
	sudo docker exec -i pg_auth_system_container psql -U dwiw auth_system_go < internal/utils/postgres/users.sql
pgMigrateEmail:
	sudo docker exec -i pg_auth_system_container psql -U dwiw auth_system_go < internal/utils/postgres/emails.sql
pgMigrateDrop:
	sudo docker exec -i pg_auth_system_container psql -U dwiw auth_system_go < internal/utils/postgres/drop_all_tables.sql
pgMigratePrivKey:
	sudo docker exec -i pg_auth_system_container psql -U dwiw auth_system_go < internal/utils/postgres/priv_key.sql
pgInsertMaritalStatus:
	sudo docker exec -i pg_auth_system_container psql -U dwiw auth_system_go < internal/utils/postgres/insert_marital_status.sql