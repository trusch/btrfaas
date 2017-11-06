FROM golang:1.9 as builder
COPY * /go/src/native-go-echo/
RUN go get -d -v native-go-echo
RUN cd /go/src/native-go-echo && CGO_ENABLED=0 go build -v -a -ldflags '-extldflags "-static"' .

FROM alpine
COPY --from=builder /go/src/native-go-echo/native-go-echo /bin/native-go-echo
CMD ["/bin/native-go-echo"]
