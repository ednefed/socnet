FROM golang:1.22.3-alpine3.20 as build

WORKDIR /app

COPY go.* .
RUN go mod download
COPY *.go .
RUN go build -o api .

FROM alpine:3.20

WORKDIR /app

COPY --from=build /app/api .

EXPOSE 8080
ENTRYPOINT [ "/app/api" ]
