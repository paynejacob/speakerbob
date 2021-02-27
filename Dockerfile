FROM node:alpine as uibuild
RUN apk add --no-cache yarn
WORKDIR /ui
COPY web/speakerbob /ui
RUN yarn install --no-lockfile --silent --cache-folder .yc \
    && yarn build

FROM golang:1.16-alpine3.13 as gobuild
RUN apk add --no-cache curl gcc musl-dev
WORKDIR /speakerbob
COPY cmd cmd
COPY pkg pkg
COPY --from=uibuild /ui/dist assets
COPY go.* ./
COPY main.go main.go
RUN go generate ./...
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o speakerbob main.go

FROM alpine:3.13
RUN apk add --no-cache ffmpeg
COPY --from=gobuild /speakerbob/speakerbob /usr/local/bin/speakerbob
EXPOSE 80
ENTRYPOINT ["/usr/local/bin/speakerbob"]