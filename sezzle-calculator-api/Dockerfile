FROM golang:1.26-alpine AS build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd

FROM alpine:3.21
RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/server ./server
COPY --from=build /build/internal/infrastructure/resource/ ./internal/infrastructure/resource/
USER 10001:10001
ENV ACTIVE_PROFILE=prod
EXPOSE 8080
ENTRYPOINT ["./server"]
