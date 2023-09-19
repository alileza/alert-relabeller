FROM golang:1.21-alpine as gobuild

COPY . /app

WORKDIR /app

RUN go build -o /bin/alert-relabeller

# final image
FROM alpine:3.17.3

WORKDIR /app

COPY --from=gobuild /bin/alert-relabeller /usr/bin/alert-relabeller

ENTRYPOINT [ "alert-relabeller" ]