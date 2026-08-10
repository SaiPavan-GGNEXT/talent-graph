# Stage 1 — build the React frontend
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY frontend/ ./
RUN npm run build

# Stage 2 — build the Go server
FROM golang:1.26-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN CGO_ENABLED=0 go build -o /bin/server ./cmd/server && \
    CGO_ENABLED=0 go build -o /bin/seed ./cmd/seed

# Stage 3 — minimal runtime
FROM alpine:3.20
RUN adduser -D app
USER app
WORKDIR /home/app
COPY --from=backend /bin/server /bin/seed ./
COPY --from=frontend /app/frontend/dist ./frontend/dist
ENV PORT=8080 STATIC_DIR=frontend/dist
EXPOSE 8080
CMD ["./server"]
