FROM golang:latest

WORKDIR /go/src/github.com/verygoodsoftwarenotvirus/tarp

ADD . .

ENTRYPOINT ["go", "test", "-v", "-cover", "-race"]
