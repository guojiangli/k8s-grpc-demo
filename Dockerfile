FROM golang AS build-env
ADD . /go/src/app
WORKDIR /go/src/app
RUN GOOS=linux GOARCH=386 go build -mod vendor cmd/svr/svr.go

FROM alpine

COPY --from=build-env /go/src/app/svr /usr/local/bin/svr

CMD [ "svr", "-port", "50009"]