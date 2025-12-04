package sidecar

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

// loadTLSCredentials loads the CA, Cert, and Key files for mTLS.
func loadTLSCredentials(caFile, certFile, keyFile string) (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed client's certificate
	pemServerCA, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(config), nil
}

// logPeerIdentityInterceptor extracts and logs the Common Name (CN) from the peer certificate.
func logPeerIdentityInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if p, ok := peer.FromContext(ctx); ok {
		if tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo); ok {
			for _, cert := range tlsInfo.State.PeerCertificates {
				log.Printf("Request from peer: %s", cert.Subject.CommonName)
			}
		}
	}
	return handler(ctx, req)
}

// streamLogPeerIdentityInterceptor extracts and logs the Common Name (CN) from the peer certificate for streams.
func streamLogPeerIdentityInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if p, ok := peer.FromContext(ss.Context()); ok {
		if tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo); ok {
			for _, cert := range tlsInfo.State.PeerCertificates {
				log.Printf("Stream from peer: %s", cert.Subject.CommonName)
			}
		}
	}
	return handler(srv, ss)
}
