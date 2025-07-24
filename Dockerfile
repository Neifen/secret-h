FROM golang:alpine AS builder

WORKDIR /
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . ./

RUN go install "github.com/a-h/templ/cmd/templ@latest"
RUN templ generate
RUN CGO_ENABLED=0  GOOS=linux go build -o bin/secret-h .

FROM scratch

COPY --from=builder bin/secret-h secret-h
COPY --from=builder assets assets

EXPOSE 8148
ENTRYPOINT ["./secret-h"]
