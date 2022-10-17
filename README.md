# Sample Golang To Do app

The application code is based on a [blog](https://blog.logrocket.com/building-simple-app-go-postgresql/) published by Emmanuel John.

## Running

```bash
export PGUSER=<pg user> # defaults to postgres
export PGPASSWORD=<pg password>
export PGSSLMODE=<ssl mode> # defaults to require
export PGHOST=<pg host> # defaults to localhost:6432
export DBNAME=<db name> # defaults to mydb
```

## Updating the app image
This applicatio is used by various demos. Build and push an updated image any time you make a change. The image is published at [https://hub.docker.com/repository/docker/syntasso/sample-todo-app](https://hub.docker.com/repository/docker/syntasso/sample-todo-app).
