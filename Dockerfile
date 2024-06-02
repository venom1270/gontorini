FROM golang:alpine3.20

RUN apk add --no-cache gcc g++ git openssh-client
RUN export CGO_ENABLED=1


WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o main .

EXPOSE 2222
CMD ["./main"]
