FROM golang:1
WORKDIR /usr/src/app
COPY . ./
RUN go build -o ackbar-adapter

EXPOSE 8080
EXPOSE 6443
ENTRYPOINT ["./ackbar-adapter"]
