build: build-server build-attacker

build-server:
	@mkdir -p bin
	@go build -o ./bin github.com/jsign/timing-attack/cmd/server

build-attacker:
	@mkdir -p bin
	@go build -o ./bin github.com/jsign/timing-attack/cmd/attacker

run-server: build-server
	@./bin/server --stddev 5 --baseLatency 15

run-attacker: build-attacker
	@./bin/attacker --debug

clean:
	@rm -rf bin

.PHONY: build build-server build-attacker run-server run-attacker clean