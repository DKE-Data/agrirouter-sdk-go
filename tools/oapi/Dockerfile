FROM golang:1.23.12-alpine3.22

# install oapi generator 
RUN go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.0

CMD ["oapi-codegen", "--version"]
