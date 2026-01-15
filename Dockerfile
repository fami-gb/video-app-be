FROM golang:1.25-alpine

RUN apk update && apk add --no-cache git

WORKDIR /app

RUN go install github.com/air-verse/air@latest

COPY . .

RUN if [ -f go.mod ]; then go mod download; fi

CMD ["air", "-c", ".air.toml"]
