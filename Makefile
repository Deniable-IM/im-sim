.PHONY: stop signal reset

stop:
	go run ./cmd/stop-sim

signal:
	go run ./cmd/signal-sim

denim:
	go run ./cmd/denim-sim

clear:
	rm -rf ./logs
	make reset

reset:
	make stop
	y | docker container prune
	docker network rm IMvlan
	docker network rm backend
	
