goose -dir ./migrations postgres "postgres://user:password@localhost:5433/checkout?sslmode=disable" status
goose -dir ./migrations postgres "postgres://user:password@localhost:5433/checkout?sslmode=disable" up

# goose -dir ./migrations postgres "postgres://user:password@localhost:5435/test?sslmode=disable" status
# goose -dir ./migrations postgres "postgres://user:password@localhost:5435/test?sslmode=disable" up