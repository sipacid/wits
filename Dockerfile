FROM golang:1.22 as build-stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY ./ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /battle-of-wits

FROM debian:bookworm-slim AS build-release-stage
RUN apt-get update && apt-get install -y ffmpeg ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=build-stage /battle-of-wits /battle-of-wits
CMD ["/battle-of-wits"]
