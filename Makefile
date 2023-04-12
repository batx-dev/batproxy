.PHONE: all
all:
	@echo "make run"

.PHONE: run
run:
	go run ./cmd/main.go \
		--dsn .batproxy/batproxy.db