# emotigo
emotigo is the main repository for our backend service.
Mainly development in golang.

## File structure
```
emotigo/
  deployment: Deployment related scripts.
  docs: Documents.
  docker: Docker and dockerfiles.
  experimental: The playground.
  module: Packages/Services.
  toolbox: Some handy scripts/tools.
  vendor: All vendor libs. Need to check into git manually.
```

## Quick Start:
```
# Launch a docker with golang development environment
cd docker/golang-dev
./build.sh
./run.sh

# Test if it works (in the docker)
cd experimental/hello
go run hello.go
```
