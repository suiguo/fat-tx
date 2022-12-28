FROM alpine:latest

WORKDIR /work

ADD ./bin/Hui-TxState/main /work/main

CMD ["./main"]

