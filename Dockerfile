FROM centurylink/ca-certs
MAINTAINER Greg Russell <gfr@google.com>
# The monitoring port
EXPOSE 9090
# The service port
EXPOSE 3001
WORKDIR /app

COPY ndt-server /app/
ENTRYPOINT ["./ndt-server"]

# Build with:
# godeps save
# docker run --rm -it -v "$GOPATH":/gopath -v "$(pwd)":/app -e "GOPATH=/gopath" -w /app golang:1.4.2 sh -c 'CGO_ENABLED=0 go build -a --installsuffix cgo --ldflags="-s" -o ndt-server'
# docker build -t m-lab/ndt-server .
# docker run --rm -it -p 3001:3001 m-lab/ndt-server
# The binary works fine on the command line, but not when
# run from the container.  Something wrong with port mapping?
