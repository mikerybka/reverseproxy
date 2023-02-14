FROM --platform=linux/amd64 golang:latest

RUN go work init
COPY . /src/github.com/mikerybka/reverseproxy
RUN go work use /src/github.com/mikerybka/reverseproxy

RUN go build -o /bin/reverseproxy github.com/mikerybka/reverseproxy

ENTRYPOINT ["/bin/reverseproxy"]
