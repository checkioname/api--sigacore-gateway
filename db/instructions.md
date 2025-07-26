##### Migrations have been running through the _golang-migrate_ library

The following command runs a migrations to our local db:

````bash 
  migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up
```
