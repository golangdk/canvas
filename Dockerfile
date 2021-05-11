FROM golang:1.16-buster AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download -x

COPY . ./
RUN go build -ldflags="-X 'main.release=`git rev-parse --short=8 HEAD`'" -o /bin/server cmd/server/*.go

FROM gcr.io/distroless/base-debian10
WORKDIR /app

COPY --from=builder /bin/server ./

CMD ["./server"]
