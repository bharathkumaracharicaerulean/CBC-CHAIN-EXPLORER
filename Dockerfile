FROM golang:1.23.9 as builder

WORKDIR /cbcscan

COPY go.mod go.sum ./
RUN go mod download
COPY . /cbcscan
WORKDIR /cbcscan/cmd
RUN go build -o cbcscan

FROM alpine:3

WORKDIR cbcscan
COPY configs configs
COPY configs/config.yaml.example configs/config.yaml

COPY --from=builder /cbcscan/cmd/cbcscan cmd/cbcscan
WORKDIR cmd
RUN apk update && apk add gcompat
ENTRYPOINT ["/cbcscan/cmd/cbcscan"]
EXPOSE 4399