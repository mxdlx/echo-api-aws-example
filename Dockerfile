FROM golang:1.10-alpine

# https://hub.docker.com/_/golang/
WORKDIR /go/src/app
COPY . .

RUN apk add --update git && rm -rf /var/cache/apk/*
RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 1323
CMD ["app"]
