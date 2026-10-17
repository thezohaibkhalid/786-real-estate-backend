FROM golang:1.27.2-alpine AS build
RUN apk add --no-cache build-base
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -trimpath -o /out/api ./cmd/api && CGO_ENABLED=1 go build -trimpath -o /out/migrate ./cmd/migrate
FROM alpine:3.22
# The dynamic encoder loads libwebp.so and libwebpdemux.so. The development
# package supplies those loader names; a static Go binary cannot load them.
RUN apk add --no-cache ca-certificates libwebp-dev && adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/ /app/
USER app
RUN /app/api --check-image-encoder
ENV HTTP_ADDR=0.0.0.0:8080
EXPOSE 8080
CMD ["/app/api"]
