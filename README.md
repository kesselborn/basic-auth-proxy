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

# usage

    $ ./basic-auth-proxy
      -addr string
            where to listen for connection (default "0.0.0.0:8443")
      -origin string
            target where you want to proxy to
      -prefix-config string
            prefix rule in the form path:username:password -- set multiple via comma separated rules (default "/::")
      -tls-cert string
            https tls certificate (only necessary when running in https mode) (default "tls.crt")
      -tls-key string
            https private key (only necessary when running in https mode) (default "tls.key")
    
    You can define multiple users and multiple paths via the 'prefix-config' parameter.
    If the path ends with a '/', this is a prefix, otherwise an exact match:
    Calling with prefix-config set to:
    
            --prefix-config "/::,/foo:foo:foo,/foo/:foo2:foo2,/foo/bar/:bar:bar,/foo/baz::"
    
    /::               -> no username / password for all paths that don't match any other rule (because "/" matches all paths)
    /foo:foo:foo      -> basic auth with foo/foo necessary for the path '/foo'
    /foo/:foo2:foo2   -> basic auth with foo2/foo2 for path '/foo/' and all its sub paths
    /foo/bar/:bar:bar -> basic auth with bar/bar for path '/foo/bar/' and all sub paths (overwrites /foo/* rule for these paths)
    /foo/baz::        -> no username/password required for the exact path '/foo/baz' (overwrites the /foo/* rule for this part)


