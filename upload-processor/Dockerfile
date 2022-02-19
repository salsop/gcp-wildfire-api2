##
## Build
##
FROM golang:1.17.6 as build

WORKDIR /app

COPY go.mod /app/go.mod
COPY go.sum /app/go.sum

RUN go mod download

COPY main.go /app/main.go

RUN go build -o /upload-processor

##
## Deploy
##

FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /upload-processor /upload-processor

EXPOSE 80

USER nonroot:nonroot

ENTRYPOINT ["/upload-processor"]
