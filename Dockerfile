FROM node:alpine as uibuild
RUN apk add --no-cache yarn
WORKDIR /ui
COPY web/speakerbob/package.json /ui/package.json
RUN yarn install --no-lockfile --silent --cache-folder .yc
COPY web/speakerbob /ui
RUN yarn build

FROM golang:1.16.6-alpine3.13 as gobuild
ARG VERSION=v100.100.100-dev
RUN apk add --no-cache curl gcc musl-dev
WORKDIR /speakerbob
COPY go.* ./
COPY main.go main.go
RUN go mod download
COPY cmd cmd
COPY pkg pkg
COPY --from=uibuild /ui/dist pkg/static/assets
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags "-X github.com/paynejacob/speakerbob/pkg/version.Version=$VERSION" -o speakerbob main.go

FROM alpine:3.13
RUN apk add --no-cache ffmpeg flite
COPY build/docker/mime.types /etc/mime.types
COPY --from=gobuild /speakerbob/speakerbob /usr/local/bin/speakerbob
VOLUME ["/data"]
EXPOSE 80
ENTRYPOINT ["/usr/local/bin/speakerbob"]
