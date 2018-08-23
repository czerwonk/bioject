[![Build Status](https://travis-ci.org/czerwonk/bioject.svg)](https://travis-ci.org/czerwonk/bioject)
[![Docker Build Statu](https://img.shields.io/docker/build/czerwonk/bioject.svg)](https://hub.docker.com/r/czerwonk/bioject/builds)
[![Go Report Card](https://goreportcard.com/badge/github.com/czerwonk/bioject)](https://goreportcard.com/report/github.com/czerwonk/bioject)

# BIOject
Route injector based on BIO routing daemon (https://github.com/bio-routing/bio-rd)

## Use cases
* automatically inject routes to mitigate DDos attacks (RTBH)

## Installation

### From Source

#### CLI Client
```bash
go get github.com/czerwonk/bioject/cmd/biojecter
```

#### Server
```bash
go get github.com/czerwonk/bioject/cmd/bioject
```

### Docker

#### Server
```bash
docker run -d --restart always --name bioject -p 179:179 -p 1337:1337 -p 6500:6500 -v /etc/bioject:/config czerwonk/bioject
```

### Configuration
```yaml
local_as: 65500
router_id: 127.0.0.1

route_filters:
  - net: "2001:678:1e0::"
    length: 48
    min: 127
    max: 128

sessions:
  - name: session1
    remote_as: 202739
    ip: 2001:678:1e0:b::1
    passive: true
```

## Third Party Components
This software uses components of the following projects
* BIO routing daemon (https://github.com/bio-routing/bio-rt)
