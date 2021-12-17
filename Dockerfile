FROM golang AS build-env
WORKDIR /go/src/app
ADD . /go/src/app
RUN GOOS=linux GOARCH=386 go build -mod vendor cmd/svr/svr.go
RUN GOOS=linux GOARCH=386 go build -mod vendor cmd/gw/gw.go

FROM alpine
COPY --from=build-env /go/src/app/svr /usr/local/bin/svr
COPY --from=build-env /go/src/app/gw /usr/local/bin/gw
COPY --from=build-env /go/src/app/start.sh /start.sh

RUN chmod +x /start.sh

CMD ["./start.sh"]