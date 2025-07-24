FROM golang:alpine AS builder

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .

RUN go install "github.com/a-h/templ/cmd/templ@latest"
RUN templ generate
RUN GOOS=linux GOARCH=arm64 go build -o bin/secret-h .

FROM scratch

COPY --from=builder bin/secret-h bin/secret-h
COPY --from=builder assets bin/assets

EXPOSE 8148
ENTRYPOINT ["bin/secret-h"]
