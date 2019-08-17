FROM node:alpine as uibuild
RUN apk add --no-cache yarn
WORKDIR /ui
COPY web /ui
RUN yarn install --no-lockfile --silent --cache-folder .yc \
    && yarn build

FROM golang:1.11-alpine3.8 as gobuild
RUN apk add --no-cache curl git gcc musl-dev
RUN curl -s https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
WORKDIR /go/src/speakerbob
COPY . .
RUN dep ensure
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o speakerbob cmd/speakerbob.go

FROM alpine:3.8
RUN apk add --no-cache ffmpeg
COPY --from=gobuild /go/src/speakerbob/speakerbob /usr/local/bin/speakerbob
COPY --from=uibuild /ui/dist /etc/speakerbob/assets
EXPOSE 80
CMD ["/usr/local/bin/speakerbob"]
