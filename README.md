# wat?

Simple proxy that can protect certain paths or prefixes via basic auth.

# build

## certificates
Generate self signed certificates by calling `./create-certs.sh`

## build go binary

    go build .

## docker image

    docker build -t basic-auth-proxy .

## run docker image

   docker run -p 8443:8443 basic-auth-proxy:latest -origin https://example.com

   curl -k https://127.0.0.1:8443
