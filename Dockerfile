FROM golang:1.18 as builder

WORKDIR /app

ENV GONOSUMDB=*
ENV GO111MODULE=on

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main

FROM centos:v8

WORKDIR /app

COPY --from=builder /app/main /app/main

EXPOSE 8000

ENTRYPOINT ["bash", "-c", "load-env /var/run/secrets/app /app/main"]