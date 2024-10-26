FROM golang:1.23.2-alpine3.20 as build

WORKDIR /app

COPY go.* .
RUN go mod download
COPY *.go .
RUN go build -o api .

FROM alpine:3.20

WORKDIR /app

COPY --from=build /app/api .
COPY *.html .

EXPOSE 8080
ENTRYPOINT [ "/app/api" ]
