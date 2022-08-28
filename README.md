# Go OAuth

## Setup

Required:

- login to [auth0.com](https://auth0.com/)
- create an API
- setup the `.env` file
- generate a certificate for the server (and CA)
  - put `ca.crt` under `client/`
  - put `server.crt` and `server.key` under `server/`

Optional:

- grant a scope in "Machine to Machine applications" to your API

## Go

### `net/http`

A Handler responds to an HTTP request.

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

The HandlerFunc type is an adapter to allow the use of ordinary functions as
HTTP handlers. If f is a function with the appropriate signature, HandlerFunc(f)
is a Handler that calls f.

```go
type HandlerFunc func(ResponseWriter, *Request)

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request)
```

ListenAndServe listens on the TCP network address addr and then calls Serve with
handler to handle requests on incoming connections. The handler is typically
nil, in which case the DefaultServeMux is used. Use Handle and HandleFunc to
register handlers in the DefaultServeMux. ListenAndServe always returns a
non-nil error.

```go
func ListenAndServe(addr string, handler Handler) error
```

### `gorilla/mux`

Router registers routes to be matched and dispatches a handler. It implements
the http.Handler interface, so it can be registered to serve requests:

```go
type Router struct {
    // Configurable Handler to be used when no route matches.
    NotFoundHandler http.Handler

    // Configurable Handler to be used when the request method does not match the route.
    MethodNotAllowedHandler http.Handler
}

// Handle registers a new route with a matcher for the URL path.
func (r *Router) Handle(path string, handler http.Handler) *Route
// HandleFunc registers a new route with a matcher for the URL path.
func (r *Router) HandleFunc(path string, f func(http.ResponseWriter, *http.Request)) *Route
// ServeHTTP dispatches the handler registered in the matched route.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request)
// Use appends a MiddlewareFunc to the chain. Middleware are executed in the order that they are applied to the Router.
func (r *Router) Use(mwf ...MiddlewareFunc)
```

MiddlewareFunc is a function which receives an http.Handler and returns another
http.Handler. Typically, the returned handler is a closure which does something
with the http.ResponseWriter and http.Request passed to it, and then calls the
handler passed as parameter to the MiddlewareFunc.

```go
type MiddlewareFunc func(http.Handler) http.Handler

// example 1: regular middleware
func Middleware1(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // do stuff before the handlers
        next.ServeHTTP(w, r)
        // do stuff after the handlers
    })
}

// example 2: middleware  with an argument
func Middleware2(s string) mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // do stuff something with the parameter
            fmt.Print(s)
            next.ServeHTTP(w, r)
        })
    }
}

// example 3: from net/http
func StripPrefix(prefix string, h Handler) Handler {
	if prefix == "" {
		return h
	}
	return HandlerFunc(func(w ResponseWriter, r *Request) {
		p := strings.TrimPrefix(r.URL.Path, prefix)
		rp := strings.TrimPrefix(r.URL.RawPath, prefix)
		if len(p) < len(r.URL.Path) && (r.URL.RawPath == "" || len(rp) < len(r.URL.RawPath)) {
			r2 := new(Request)
			*r2 = *r
			r2.URL = new(url.URL)
			*r2.URL = *r.URL
			r2.URL.Path = p
			r2.URL.RawPath = rp
			h.ServeHTTP(w, r2)
		} else {
			NotFound(w, r)
		}
	})
}
```

```go
type Route struct {
    // contains filtered or unexported fields
}

// Handler sets a handler for the route.
func (r *Route) Handler(handler http.Handler) *Route
```

## OAuth 2.0

Roles:

- user (e.g., Federico)
- application (e.g., Spotify)
- API (e.g., Google, Facebook, or GitHub)
  - authorization server
  - resource server

OAuth is an authorization framework that allow *users* to grant access to API
*server* resources to another *application* without sharing credentials.

The application has to be registered with the API service, usually providing app
name, website and a redirect URL. With registration, the application gets a
public client id and a (private) client secret. To request an access token, the
application needs to provide these credentials to the API authorization server.

All OAuth 2.0 traffic must be encryped via TLS. Grant types specify the kinds of
authorization scenarios that OAuth has been designed to handle. There are four
grant types. Scopes can be used to limit access for a specific access token.

Requesting an access token:

```bash
curl  \
    -H "Accept: application/json" \
    -H 'Content-Type: application/json' \
    -X POST https://$AUTH0_DOMAIN/oauth/token \
    --data @<(cat <<EOF
{
    "client_id": "$OAUTH_CLIENT_ID",
    "client_secret": "$OAUTH_CLIENT_SECRET",
    "audience": "$AUTH0_AUDIENCE",
    "grant_type": "client_credentials"
}
EOF
)
```

RS256 works by using a private/public key pair, tokens can be verified against
the public key for your Auth0 account. This public key is accessible at
`https://$AUTH0_DOMAIN/.well-known/jwks.json`.

## Certificates

Generate CA's key pair and certificate.

```bash
openssl genrsa -aes256 -out ca.key 4096
openssl req -x509 -new -nodes -key ca.key -sha512 -days 1825 -out ca.crt
```

Install it on your Windows machine with PowerShell with admin permissions.

```pwsh
Import-Certificate -FilePath "\\wsl$\Ubuntu\home\fedev\certs\ca.crt" -CertStoreLocation Cert:\LocalMachine\Root
```

Or in Linux:

```bash
sudo cp ~/certs/ca.crt /usr/local/share/ca-certificates/ca.crt
sudo update-ca-certificates
```

Generate the certificate of a specific server/domain.

```bash
# generate key pair
openssl genrsa -out fgserver.test.key 2048
# create CSR (certificate signing request)
openssl req -new -key fgserver.test.key -out fgserver.test.csr
# create X509 V3 certificate extension config file
cat > fgserver.test.ext << EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names
[alt_names]
DNS.1 = fgserver.test
DNS.2 = localhost
EOF
# sign the certificate + the config file
openssl x509 -req -CA ca.crt -CAkey ca.key \
    -extfile fgserver.test.ext \
    -in fgserver.test.csr \
    -out fgserver.test.crt \
    -CAcreateserial -days 825 -sha256
```

## TLS

On the client side, use:

```bash
curl -v -H "Authorization: Bearer ${TOKEN}" https://localhost:$PORT/api/v1/products --cacert ca.crt
```

On the server side, the `ListenAndServeTLS()` function documentation states:

> If the certificate is signed by a certificate authority, the certFile should
> be the concatenation of the server's certificate, any intermediates, and the
> CA's certificate.

## Docker

Build image:

```bash
cd server
docker build -t gosrv .
```

Run the container:

```bash
docker run --rm gosrv
# or
docker run --network="host" --rm gosrv
```

Useful:

- `docker inspect bridge | grep -B 3 -A 2 "IPv4"`
- `curl -k -D /dev/stderr -w "\n" --cacert ca.crt -H "Authorization: Bearer
  ${TOKEN}" -X GET https://$ADDR:$PORT/api/v1/products`

## Useful Resources

- [Understanding HandlerFunc on
  StackOverflow](https://stackoverflow.com/questions/53678633/understanding-the-http-handlerfunc-wrapper-technique-in-go)
- [Auth0
  sample](https://github.com/auth0-samples/auth0-golang-api-samples/tree/master/01-Authorization-RS256)
- [Example
  scenario](https://auth0.com/docs/get-started/apis/scopes/api-scopes#example-an-api-called-by-a-back-end-service)
- [Local CA for
  development](https://deliciousbrains.com/ssl-certificate-authority-for-local-https-development/)
- [Go SSL server with perfect
  score](https://blog.bracebin.com/achieving-perfect-ssl-labs-score-with-go)

## TODO

Nice to have:

- validate access token only for some paths of the API
- CORS
- graceful degradation
- mTLS
  - https://smallstep.com/hello-mtls/doc/server/go
  - https://smallstep.com/hello-mtls/doc/combined/go/curl
