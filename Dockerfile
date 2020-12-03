FROM golang:1.15-alpine
WORKDIR /app
COPY . .
RUN go mod vendor
RUN CGO_ENABLED=0 go build -mod vendor
