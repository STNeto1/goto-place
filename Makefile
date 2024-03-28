seed:
	rm db.sqlite
	sqlite3 db.sqlite < setup.sql

build-app:
	cd ui && \
	bun install && \
    bun run build && \
    cd .. && \
    go build -o bin/server
