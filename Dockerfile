FROM golang:1.11-alpine

RUN wget -q https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 -O /usr/local/bin/dep \
    && chmod +x /usr/local/bin/dep

RUN mkdir -p $GOPATH/src/stevenolen/504
COPY . $GOPATH/src/stevenolen/504
WORKDIR $GOPATH/src/stevenolen/504
RUN dep ensure
RUN go install

EXPOSE 8080
ENTRYPOINT ["504"]
