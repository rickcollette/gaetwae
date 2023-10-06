package tls

import (
	"crypto/tls"
	"net/http"
)

// StartHTTPSServer starts an HTTPS server with the given handlers, certificate, and key.
// It returns an error if there's an issue starting the server or loading the certificate.
func StartHTTPSServer(handlers http.Handler, certPath, keyPath string) error {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	server := &http.Server{
		Addr:      ":443",
		Handler:   handlers,
		TLSConfig: tlsConfig,
	}

	return server.ListenAndServeTLS("", "")
}
