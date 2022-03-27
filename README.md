# Port Finder

Port scanner performs search over range of ports, try to connect and detect web services

## Build from source

Clone repository and then

```
make && sudo make install
```

Also you may use [sokil/port-finder](https://hub.docker.com/r/sokil/port-finder) docekr image

## Useage

Scan IP address on all renge of ports:

```
docker run --rm sokil/port-finder -ip 8.8.8.8
```

Scan CIDR IP range:

```
docker run --rm sokil/port-finder -ip 8.8.8.8/24
```

Scan IP range:
```
docker run --rm sokil/port-finder -ip 8.8.8.8-8.8.200
```

Scan limited range of ports:

```
docker run --rm sokil/port-finder -ip 8.8.8.8 -portRange 70-90
```

Scan ports from defined to last

```
docker run --rm sokil/port-finder -ip 8.8.8.8 -portRange 70-
```

Scan ports from first to defined:

```
docker run --rm sokil/port-finder -ip 8.8.8.8 -portRange -90
```

Define concurrency (how many ports to check one time):

```
docker run --rm sokil/port-finder -ip 8.8.8.8 -concurrent 200
```

Show check errors:

```
docker run --rm sokil/port-finder -ip 8.8.8.8 -verbose
```
