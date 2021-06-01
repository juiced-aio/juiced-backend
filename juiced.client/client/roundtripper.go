package client

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"

	utls "github.com/Titanium-ctrl/utls"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.client/http2"
)

var errProtocolNegotiated = errors.New("protocol negotiated")

type DialFunc func(context.Context, string, string) (net.Conn, error)

type roundTripper struct {
	sync.Mutex

	InsecureSkipVerify bool

	Network string

	clientHelloId utls.ClientHelloID

	cachedConnections map[string]net.Conn
	cachedTransports  map[string]http.RoundTripper
	DebugCountBytes   func(uint8, uint)
	dialer            proxy.ContextDialer
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	addr := rt.getDialTLSAddr(req)
	if _, ok := rt.cachedTransports[addr]; !ok {
		if err := rt.getTransport(req, addr); err != nil {
			return nil, err
		}
	}

	resp, err := rt.cachedTransports[addr].RoundTrip(req)

	return resp, err
}

func (rt *roundTripper) getTransport(req *http.Request, addr string) error {
	switch strings.ToLower(req.URL.Scheme) {
	case "http":
		rt.cachedTransports[addr] = &http.Transport{DialContext: rt.dialer.DialContext}
		return nil
	case "https":
	default:
		return fmt.Errorf("invalid URL scheme: [%v]", req.URL.Scheme)
	}

	_, err := rt.dialTLS(context.Background(), "tcp", addr)
	switch err {
	case errProtocolNegotiated:
	case nil:
		// Should never happen.
		panic("dialTLS returned no error when determining cachedTransports")
	default:
		return err
	}

	return nil
}

func (rt *roundTripper) dialTLS(ctx context.Context, network, addr string) (net.Conn, error) {
	rt.Lock()
	defer rt.Unlock()

	// If we have the connection from when we determined the HTTPS
	// cachedTransports to use, return that.
	if conn := rt.cachedConnections[addr]; conn != nil {
		delete(rt.cachedConnections, addr)
		return conn, nil
	}

	/* sslConn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
	for _, cert := range sslConn.ConnectionState().PeerCertificates {
		fmt.Println(cert.Issuer)
	} */

	rawConn, err := rt.dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	var host string
	if host, _, err = net.SplitHostPort(addr); err != nil {
		host = addr
	}

	conn := utls.UClient(rawConn, &utls.Config{
		ServerName:         host,
		InsecureSkipVerify: true,
	}, rt.clientHelloId)
	if err = conn.Handshake(); err != nil {
		conn.Close()
		return nil, err
	}

	for _, cert := range conn.ConnectionState().PeerCertificates {
		stringedCert := strings.ToLower(fmt.Sprint(cert.Issuer))
		if strings.Contains(stringedCert, "charles") || strings.Contains(stringedCert, "fiddler") || strings.Contains(stringedCert, "mitm") {
			conn.Close()
			return nil, errors.New("bad proxy")
		}
	}

	if rt.cachedTransports[addr] != nil {
		return conn, nil
	}

	// No http.Transport constructed yet, create one based on the results
	// of ALPN.
	switch conn.ConnectionState().NegotiatedProtocol {
	case http2.NextProtoTLS:
		// The remote peer is speaking HTTP 2 + TLS.
		rt.cachedTransports[addr] = &http2.Transport{
			DialTLS:                rt.dialTLSHTTP2,
			DisableCompression:     false,
			MaxHeaderListSize:      262144,
			InitialWindowSize:      6291456,
			InitialHeaderTableSize: 65536,
			PushHandler:            newPushHandler(),
			DebugCountBytes:        rt.DebugCountBytes,
		}
	default:
		// Assume the remote peer is speaking HTTP 1.x + TLS.
		rt.cachedTransports[addr] = &http.Transport{DialTLSContext: rt.dialTLS, DisableCompression: false, DisableKeepAlives: false, MaxIdleConns: 0}
	}

	// Stash the connection just established for use servicing the
	// actual request (should be near-immediate).
	rt.cachedConnections[addr] = conn

	return nil, errProtocolNegotiated
}

func newPushHandler() *PushHandler {
	return &PushHandler{
		done: make(chan struct{}),
	}
}

type PushHandler struct {
	promise          *http.Request
	origReqURL       *url.URL
	origReqRawHeader http.RawHeader
	origReqHeader    http.Header
	push             *http.Response
	pushErr          error
	done             chan struct{}
}

func (ph *PushHandler) HandlePush(r *http2.PushedRequest) {
	ph.promise = r.Promise
	ph.origReqURL = r.OriginalRequestURL
	ph.origReqRawHeader = r.OriginalRequestRawHeader
	ph.origReqHeader = r.OriginalRequestHeader
	ph.push, ph.pushErr = r.ReadResponse(r.Promise.Context())
	if ph.pushErr != nil {
		defer ph.push.Body.Close()
	}
	if ph.push != nil {
		ioutil.ReadAll(ph.push.Body)
		time.Sleep(1000 * time.Millisecond)
		defer ph.push.Body.Close()
	}
}

func (rt *roundTripper) dialTLSHTTP2(network, addr string, _ *tls.Config) (net.Conn, error) {
	return rt.dialTLS(context.Background(), network, addr)
}

func (rt *roundTripper) getDialTLSAddr(req *http.Request) string {
	host, port, err := net.SplitHostPort(req.URL.Host)
	if err == nil {
		return net.JoinHostPort(host, port)
	}
	return net.JoinHostPort(req.URL.Host, "443") // we can assume port is 443 at this point
}

func newRoundTripper(clientHello utls.ClientHelloID, dialer ...proxy.ContextDialer) http.RoundTripper {
	if len(dialer) > 0 {
		return &roundTripper{
			dialer:             dialer[0],
			clientHelloId:      clientHello,
			InsecureSkipVerify: true, //os.Getenv("INSECURE_SKIP_VERIFY") == "1",
			cachedTransports:   make(map[string]http.RoundTripper),
			cachedConnections:  make(map[string]net.Conn),
		}
	} else {
		return &roundTripper{
			dialer:             proxy.Direct,
			clientHelloId:      clientHello,
			InsecureSkipVerify: true, //os.Getenv("INSECURE_SKIP_VERIFY") == "1",
			cachedTransports:   make(map[string]http.RoundTripper),
			cachedConnections:  make(map[string]net.Conn),
		}
	}
}
