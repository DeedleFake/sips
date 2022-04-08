FROM golang:1.18-alpine AS build

WORKDIR /build
COPY . /build

RUN apk add git

ENV CGO_ENABLED 0
RUN go build -o sips ./cmd/sips
RUN go build -o sipsctl ./cmd/sipsctl

FROM scratch

COPY --from=build /build/sips /sips
COPY --from=build /build/sipsctl /sipsctl

ENV PATH /

EXPOSE 8080

CMD ["sips"]
