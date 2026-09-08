You are using postgres with this project. You are using golang-migrate to create and update your database.

The command to migrate up (if you are starting raw, also the command to create) is 

migrate \
  -path ./migrations \
  -database "postgres://postgres:password@localhost:5432/uptime_monitor?sslmode=disable" \
  up

  Run this command ^ from your root directory.

## COMMAND TO DELETE AFTER REBOOT
# sudo systemctl stop postgresql

## COMMAND TO START CONTAINER
# docker start uptime-postgres

## COMMAND TO START POSTGRES FROM SCRATCH IF CONTAINER DOES NOT EXIST
# docker run --name uptime-postgres \
#   -e POSTGRES_USER=postgres \
#   -e POSTGRES_PASSWORD=password \
#   -e POSTGRES_DB=uptime_monitor \
#   -p 5432:5432 \
#   -d postgres:14.24-alpine3.23
