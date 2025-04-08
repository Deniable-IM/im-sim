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
	sudo docker container prune
	sudo docker network rm IMvlan
	sudo docker network rm backend
	
