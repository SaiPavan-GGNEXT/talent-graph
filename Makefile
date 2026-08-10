.PHONY: dev build seed run frontend clean

# Build everything (frontend + server binary)
build: frontend
	go build -o bin/server ./cmd/server
	go build -o bin/seed ./cmd/seed

frontend:
	cd frontend && npm install --no-audit --no-fund && npm run build

# Load the demo dataset (wipes existing data)
seed:
	go run ./cmd/seed

# Run the API + built frontend on :8080
run:
	go run ./cmd/server

# Frontend dev server with API proxy (run `make run` in another terminal)
dev:
	cd frontend && npm run dev

clean:
	rm -rf bin frontend/dist
