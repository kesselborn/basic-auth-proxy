package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

// PrefixConfig stores a prefix configuration. Prefix is the path
// prefix which should be guarded by username/password
type PrefixConfig struct {
	prefix    string
	username  string
	password  string
	protected bool
}

func (pc PrefixConfig) String() string {
	if pc.protected {
		return fmt.Sprintf("%-30s (user: %s, password: ******)", pc.prefix, pc.username)
	}

	return fmt.Sprintf("%-30s (no username/password required)", pc.prefix)
}

// NewPrefixConfig returns a config from a string with format '<prefix>:<username>:<password>'
func NewPrefixConfig(config string) PrefixConfig {
	tokens := strings.Split(config, ":")
	if len(tokens) != 3 {
		log.Fatalf("error: prefix config format must be '<prefix>:<username>:<password>', was: %s\n", config)
	}

	return PrefixConfig{protected: tokens[2] != "", prefix: tokens[0], username: tokens[1], password: tokens[2]}
}

// Serve returns a handler for given reverse proxy with given prefixConfig
func Serve(prefixConfig PrefixConfig, proxy http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, basicAuthProvided := r.BasicAuth()
		r.SetBasicAuth("", "") // don't pass the basic auth credentials to origin

		if prefixConfig.protected {
			if !basicAuthProvided {
				w.Header().Add("WWW-Authenticate", "Basic realm=user:"+prefixConfig.username)
				http.Error(w, "authentication required", 401)
				return
			}

			if username != prefixConfig.username || password != prefixConfig.password {
				http.Error(w, "wrong username / password", 403)
				return
			}
		}

		proxy(w, r)
	}
}

func main() {
	addr := flag.String("addr", "0.0.0.0:8443", "where to listen for connection")
	originURL := flag.String("origin", "", "target where you want to proxy to")
	prefixConfig := flag.String("prefix-config", "/::", "prefix rule in the form path:username:password -- set multiple via comma separated rules")
	tlsCert := flag.String("tls-cert", "tls.crt", "https tls certificate (only necessary when running in https mode)")
	tlsKey := flag.String("tls-key", "tls.key", "https private key (only necessary when running in https mode)")
	httpMode := flag.Bool("run-in-http-mode-although-i-know-i-shouldnt-do-this", false, "run in http mode -- you should never do this unless testing locally")

	flag.Parse()

	if *originURL == "" {
		flag.PrintDefaults()
		fmt.Print(`
You can define multiple users and multiple paths via the 'prefix-config' parameter.
If the path ends with a '/', this is a prefix, otherwise an exact match:
Calling with prefix-config set to:

	--prefix-config "/::,/foo:foo:foo,/foo/:foo2:foo2,/foo/bar/:bar:bar,/foo/baz::"

/::               -> no username / password for all paths that don't match any other rule (because "/" matches all paths)
/foo:foo:foo      -> basic auth with foo/foo necessary for the path '/foo'
/foo/:foo2:foo2   -> basic auth with foo2/foo2 for path '/foo/' and all its sub paths
/foo/bar/:bar:bar -> basic auth with bar/bar for path '/foo/bar/' and all sub paths (overwrites /foo/* rule for these paths)
/foo/baz::        -> no username/password required for the exact path '/foo/baz' (overwrites the /foo/* rule for this part)

`)
		os.Exit(1)
	}

	origin, err := url.Parse(*originURL)
	if err != nil {
		log.Fatalf("error reading origin url: %s", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(origin)
	proxy.ErrorLog = log.New(os.Stderr, "", 0)
	proxy.Director = func(r *http.Request) {
		r.Host = origin.Host
		r.URL.Host = origin.Host
		r.URL.Scheme = origin.Scheme
	}

	for _, config := range strings.Split(*prefixConfig, ",") {
		prefixConfig := NewPrefixConfig(config)
		log.Printf("creating handler for: %s\n", prefixConfig)
		http.HandleFunc(prefixConfig.prefix, Serve(NewPrefixConfig(config), proxy.ServeHTTP))
	}

	log.Printf("listening on %s, proxying to: %s\n", *addr, origin)
	if *httpMode {
		log.Printf("RUNNING IN HTTP-MODE -- DON'T DO THIS!")
		log.Fatalf("error: %s", http.ListenAndServe(*addr, nil))
	} else {
		log.Fatalf("error: %s", http.ListenAndServeTLS(*addr, *tlsCert, *tlsKey, nil))
	}
}
