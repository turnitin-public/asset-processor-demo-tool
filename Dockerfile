FROM golang:1.23.6

RUN apt-get update && apt-get install -y libdlib-dev libblas-dev libatlas-base-dev liblapack-dev libjpeg62-turbo-dev

WORKDIR /go/src/app
COPY ./src/go.* .
ENV GO111MODULE=on

RUN go mod download
COPY ./src .
RUN go mod vendor -v
RUN go get -d -v ./...
RUN go install -v ./...

CMD ["bash", "-c", "go run *.go"]
