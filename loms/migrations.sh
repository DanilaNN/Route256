goose -dir ./migrations postgres "postgres://user:password@localhost:5434/loms?sslmode=disable" status
goose -dir ./migrations postgres "postgres://user:password@localhost:5434/loms?sslmode=disable" up