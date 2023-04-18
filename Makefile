.PHONE: all
all:
	@echo "make run"

.PHONE: run
run:
	go run ./cmd run \
		--dsn .batproxy/batproxy.db