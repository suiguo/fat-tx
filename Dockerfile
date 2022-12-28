FROM alpine:latest

WORKDIR /work

ADD ./bin/Hui-TxState/main /work/main
ADD config.yaml /work/

CMD ["./main"]

