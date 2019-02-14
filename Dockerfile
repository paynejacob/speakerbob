FROM node:alpine as uibuild
WORKDIR /app
COPY web /app
RUN apk add --no-cache yarn
RUN yarn install --no-lockfile --silent --cache-folder .yc \
    && yarn build

FROM golang:1.11-alpine3.8 as gobuild
WORKDIR /go/src/speakerbob
RUN apk add --no-cache curl git gcc musl-dev
RUN curl -s https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
COPY . .
RUN dep ensure
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o speakerbob cmd/speakerbob.go

FROM alpine:3.8
RUN apk add --no-cache ffmpeg
VOLUME ["/etc/speakerbob", "/etc/speakerbob/sounds"]
WORKDIR /root/
COPY --from=gobuild /go/src/speakerbob/speakerbob /usr/local/bin/speakerbob
COPY --from=uibuild /go/src/speakerbob/web/dist/* /etc/speakerbob/assets
EXPOSE 80
ENTRYPOINT ["/usr/local/bin/speakerbob"]
