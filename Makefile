shell: start
	docker compose exec -ti godev bash

start:
	docker compose up -d

build-lx: start dist
	docker compose exec godev go build -o dist/transmission-rutracker-seed main.go

build-win: start dist
	docker compose exec -e GOOS=windows godev go build -o dist/transmission-rutracker-seed.exe main.go

build: build-lx build-win

release: release-lx release-win

release-lx: build-lx
	docker compose exec godev strip dist/transmission-rutracker-seed

release-win: build-win
	docker compose exec godev strip dist/transmission-rutracker-seed.exe

dist:
	mkdir -p dist
