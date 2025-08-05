FROM golang:1.23.11-alpine3.22 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

WORKDIR /app/internal/tests/test_server

RUN CGO_ENABLED=0 GOOS=linux go build -o /build-bin ./cmd && chmod +x /build-bin

FROM gcr.io/distroless/base-debian12:nonroot@sha256:201ef9125ff3f55fda8e0697eff0b3ce9078366503ef066653635a3ac3ed9c26

WORKDIR /

COPY --from=build-stage --chown=nonroot:nonroot /build-bin /build-bin

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/build-bin"]