FROM golang

WORKDIR /app

COPY . .

RUN go get -v ./...

RUN make build_engine

EXPOSE 8080

CMD ["./bin/app"]
