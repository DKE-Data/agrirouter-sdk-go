FROM golang:1.23.12-alpine3.22 AS build-stage

# install oapi generator 
RUN go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.0

FROM gcr.io/distroless/base-debian12:nonroot@sha256:201ef9125ff3f55fda8e0697eff0b3ce9078366503ef066653635a3ac3ed9c26

WORKDIR /

COPY --from=build-stage --chown=nonroot:nonroot /go/bin/oapi-codegen /usr/local/bin/oapi-codegen

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["oapi-codegen", "--version"]
