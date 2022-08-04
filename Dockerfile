#syntax=docker/dockerfile:1

FROM golang:1.17.1

RUN mkdir /app
WORKDIR /app

COPY . .

RUN go get -u golang.org/x/sys
# Download all the dependencies
RUN go mod download

# Build the Go app
RUN go build -o /server
EXPOSE 8080

ENTRYPOINT [ "/server", "-mode","prod" ]
