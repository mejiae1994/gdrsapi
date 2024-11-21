.PHONY: r b
r:
	go run cmd/api/main.go

b:
	go build -o gdrsapi cmd/api/main.go