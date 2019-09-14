build: ensure-bin build-server build-attacker

build-server:
	@go build -o ./bin github.com/jsign/timing-attack/cmd/server

build-attacker:
	@go build -o ./bin github.com/jsign/timing-attack/cmd/attacker

run-server: build-server
	@./bin/server --stddev 5 --baseLatency 15 --debug

run-attacker: build-attacker
	@./bin/attacker --debug

ensure-bin:
	@mkdir -p bin

clean:
	@rm -rf bin

.PHONY: build build-server build-attacker run-server run-attacker ensure-bin clean