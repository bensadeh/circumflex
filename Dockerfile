FROM docker.io/golang:1.19.3 as clx-builder

WORKDIR /go/src/cirumflex

COPY . .

RUN CGO_ENABLED=0 go build && mv clx /go/bin

FROM alpine as less-builder

RUN apk add curl tar make gcc g++ ncurses-dev

RUN curl -LJO https://www.greenwoodsoftware.com/less/less-608.tar.gz

RUN tar --no-same-owner -xvf less-608.tar.gz

WORKDIR /less-608

RUN /less-608/configure && make install

FROM alpine

RUN apk add ncurses-dev

COPY --from=less-builder /usr/local/bin/less /usr/local/bin/less

COPY --from=clx-builder /go/bin/clx /usr/local/bin

ENTRYPOINT ["/usr/local/bin/clx"]
