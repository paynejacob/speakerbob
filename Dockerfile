FROM node:alpine as uibuild
CMD ["echo", "TODO build the ui"]

FROM golang:1.10 as gobuild
WORKDIR /go/src/speakerbob
RUN curl -s https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
COPY . .
RUN dep ensure
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o speakerbob cmd/speakerbob.go

FROM alpine:latest
WORKDIR /root/
COPY --from=gobuild /go/src/speakerbob/speakerbob /usr/local/bin/speakerbob
CMD ["echo", "TODO copy the ui"]
EXPOSE 80
CMD ["speakerbob"]
