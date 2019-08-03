FROM golang:alpine AS build-env
ENV GO111MODULE=on

WORKDIR /app

RUN apk update
RUN apk add git
RUN apk add gcc libc-dev

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY *.go ./
COPY db/ ./db/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./warbler

# ----------------------------------------

FROM clojure:lein AS frontend-env

WORKDIR /app
COPY frontend/ .
RUN ./futura_download.sh
RUN lein clean
RUN lein cljsbuild once min

# ----------------------------------------

FROM alpine
RUN apk update
RUN apk add ca-certificates
RUN rm -rf /var/cache/apk/*
RUN apk add postgresql-client
RUN apk add ffmpeg
WORKDIR /app
COPY --from=build-env /app/warbler /app/
COPY --from=frontend-env /app/resources/public /app/frontend/resources/public

EXPOSE 8080
ENTRYPOINT ["/app/warbler"]
