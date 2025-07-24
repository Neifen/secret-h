FROM golang:alpine AS builder

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .

RUN go tool templ generate
RUN GOOS=linux GOARCH=arm64 go build -o bin/secret-h .

FROM scratch

COPY --from=builder bin/secret-h bin/secret-h

EXPOSE 8148
ENTRYPOINT ["bin/secret-h"]
