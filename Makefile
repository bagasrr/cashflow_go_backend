# Command untuk run normal
run:
	swag init
	go run main.go

# Command untuk build binary (opsional)
build:
	swag init
	go build -o bin/cashflow main.go