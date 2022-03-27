# Port Finder


[![docker](https://img.shields.io/docker/pulls/sokil/port-finder.svg?style=flat)](https://hub.docker.com/r/sokil/port-finder/)

Port scanner performs search over range of ports, try to connect and detect web services

## Build from source

Clone repository and then

```
make && sudo make install
```

Also you may use [sokil/port-finder](https://hub.docker.com/r/sokil/port-finder) docekr image

## Useage

## IP Range

Scan IP address on all renge of ports:

```
docker run --rm sokil/port-finder -ipRange 8.8.8.8
```

Scan CIDT IP range:

```
docker run --rm sokil/port-finder -ipRange 8.8.8.8/24
```

Scan IP range:
```
docker run --rm sokil/port-finder -ipRange 8.8.8.8-8.8.8.200
```

## Port range

Scan limited range of ports:

```
docker run --rm sokil/port-finder -ipRange 8.8.8.8 -portRange 70-90
```

Scan ports from defined to last

```
docker run --rm sokil/port-finder -ipRange 8.8.8.8 -portRange 70-
```

Scan ports from first to defined:

```
docker run --rm sokil/port-finder -ipRange 8.8.8.8 -portRange -90
```

## Probes

Tools may perform probes for different services. Currently tool supports next probes:

* http
* https
* memcached

By default all probes performerd. To specify concrete probes use `-probe` option, delimited by space:

```
docker run --rm sokil/port-finder -ipRange 8.8.8.8 -probes="https memcached"
```

## Concurrency

Define concurrency (how many ports to check one time):

```
docker run --rm sokil/port-finder -ipRange 8.8.8.8 -concurrent 200
```

## Debug

Show check errors:

```
docker run --rm sokil/port-finder -ipRange 8.8.8.8 -verbose
```
