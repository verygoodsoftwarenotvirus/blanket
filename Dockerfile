# build stage
FROM golang:alpine AS build-stage
WORKDIR /go/src/github.com/verygoodsoftwarenotvirus/tarp

ADD . .
RUN go build -o /tarp

# final stage
FROM alpine:latest

COPY --from=build-stage /tarp /tarp

ENTRYPOINT ["/tarp"]
