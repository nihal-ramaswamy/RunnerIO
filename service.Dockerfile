FROM golang

WORKDIR /app

COPY . .

RUN go get -v ./...

RUN make build_service

EXPOSE 8080

CMD ["./bin/app"]
