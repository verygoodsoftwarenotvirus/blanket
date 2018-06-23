# build stage
FROM golang:alpine AS build-stage
WORKDIR /go/src/github.com/verygoodsoftwarenotvirus/blanket

ADD . .
RUN go build -o /blanket github.com/verygoodsoftwarenotvirus/blanket/cmd/blanket

# final stage
FROM alpine:latest

COPY --from=build-stage /blanket /blanket

ENTRYPOINT ["/blanket"]
