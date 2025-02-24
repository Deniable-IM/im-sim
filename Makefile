.PHONY: stop signal reset

stop:
	go run ./cmd/stop-sim

signal:
	go run ./cmd/signal-sim

reset:
	make stop 
	sudo docker network rm IMvlan
	
