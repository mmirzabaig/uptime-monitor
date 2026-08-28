You are using postgres with this project. You are using golang-migrate to create and update your database.

The command to migrate up (if you are starting raw, also the command to create) is 

migrate \
  -path ./migrations \
  -database "postgres://postgres:password@localhost:5432/uptime_monitor?sslmode=disable" \
  up

  Run this command ^ from your root directory.