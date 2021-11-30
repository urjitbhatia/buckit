FROM golang:1.17-alpine as builder

WORKDIR /app
ADD . /app
RUN go build -o buckit

FROM alpine

COPY --from=builder /app/buckit /usr/local/bin/
ENTRYPOINT ["buckit"]
