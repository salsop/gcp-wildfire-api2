##
## Build
##
FROM golang:1.17.6 as build

WORKDIR /app

COPY go.mod /app/go.mod
COPY go.sum /app/go.sum

RUN go mod download

COPY main.go /app/main.go

RUN go build -o /web-file-upload

##
## Deploy
##

FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY upload-page.html /upload-page.html
COPY response-page.html /response-page.html
COPY --from=build /web-file-upload /web-file-upload

EXPOSE 80

USER nonroot:nonroot

ENTRYPOINT ["/web-file-upload"]
