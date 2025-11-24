FROM golang:1.24 AS build-stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY ./ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /wits

FROM debian:bookworm-slim AS build-release-stage
RUN apt-get update && apt-get install -y ffmpeg ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=build-stage /wits /wits
CMD ["/wits"]
