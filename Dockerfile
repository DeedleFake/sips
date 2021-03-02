FROM golang:1.16 AS build

WORKDIR /go/src
COPY go ./go
COPY main.go .

ENV CGO_ENABLED=0
RUN go get -d -v ./...
RUN go build -a -installsuffix cgo -o sips .

FROM scratch AS runtime

COPY --from=build /go/src/sips ./

EXPOSE 8080/tcp
ENTRYPOINT ["./sips"]
