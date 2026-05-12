# Simple Makefile for a Go project

iro:
	@echo "Starting Iro..."
	cd apps/iro && make watch 

build-iro:
	@echo "Building Iro..."
	cd apps/iro && make build-prod

kazu:
	@echo "Starting Kazu..."
	cd apps/kazu && make watch 

kazu-migrate:
	@echo "Running Kazu migrations..."
	cd apps/kazu && make migrate 

build-kazu:
	@echo "Building Kazu..."
	cd apps/kazu && make build-prod

kusari:
	@echo "Starting Kusari..."
	cd apps/kusari && make watch 

kusari-migrate:
	@echo "Running Kusari migrations..."
	cd apps/kusari && make migrate 

build-kusari:
	@echo "Building Kusari..."
	cd apps/kusari && make build-prod

hoshi:
	@echo "Starting Hoshi..."
	cd apps/hoshi && make watch

hoshi-migrate:
	@echo "Running Hoshi migrations..."
	cd apps/hoshi && make migrate

build-hoshi:
	@echo "Building Hoshi..."
	cd apps/hoshi && make build-prod


.PHONY: iro kazu kusari hoshi
