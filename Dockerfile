FROM alpine:3.13.2 as build

RUN \
    # install requirements
    apk update && \
    apk add --no-cache \
        ca-certificates \
        make \
        bash \
        wget \
        git \
        curl \
        go

COPY . /build

# make and install source
RUN cd /build && make build && make install

FROM alpine:3.13.2 as run

COPY --from=build /usr/local/bin/port-finder /usr/local/bin/port-finder

# start service
ENTRYPOINT ["/usr/local/bin/port-finder"]
