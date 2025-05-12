FROM golang 

WORKDIR app

COPY . .

RUN go get -d -v ./...

RUN make build

EXPOSE 8080

CMD ["./bin/app"]
