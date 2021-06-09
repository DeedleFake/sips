FROM golang:1.16-alpine AS build

WORKDIR /build
COPY . /build

ENV CGO_ENABLED 0
RUN go build -o sips ./cmd/sips
RUN go build -o sipsctl ./cmd/sipsctl

FROM scratch

COPY --from=build /build/sips /sips
COPY --from=build /build/sipsctl /sipsctl

ENV PATH /
EXPOSE 8080
VOLUME /data

ENTRYPOINT ["/sips", "-db", "/data/database.db"]
