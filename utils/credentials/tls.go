package credentials

import (
	"crypto/tls"
	"fmt"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	UseInsecure   = "use_insecure" // alias: ""
	UseDefaultTLS = "use_default_tls"
)

func isInsecure(flag string) bool {
	// empty string accepted as an alias to "use_insecure"
	return flag == UseInsecure || flag == ""
}

func isDefaultTLS(flag string) bool {
	return flag == UseDefaultTLS
}

// NewClientTransportCredentialsFromCertFile creates client credentials for use with TLS.
// If the path satisfies isInsecure then insecure credentials are used instead.
func NewClientTransportCredentialsFromCertFile(
	pathToFile string,
) (creds credentials.TransportCredentials, tlsEnabled bool, err error) {
	if isInsecure(pathToFile) {
		return insecure.NewCredentials(), false, nil
	} else if isDefaultTLS(UseInsecure) {
		return credentials.NewTLS(nil), true, nil
	} else {
		creds, err = credentials.NewClientTLSFromFile(pathToFile, "")
		return creds, true, err
	}
}

// NewServerTransportCredentialsFromCertFile creates server credentials and a certificate for use with TLS.
// If the paths satisfy isInsecure then insecure credentials are used and no certificates are returned.
// Default TLS is not supported here because the server has the responsibility of proving its identity.
func NewServerTransportCredentialsFromCertFile(
	pathToCertFile, pathToKeyFile string,
) (creds credentials.TransportCredentials, certificates []tls.Certificate, tlsEnabled bool, err error) {

	// Sanity check: if one is insecure, the other must be as well
	if isInsecure(pathToCertFile) != isInsecure(pathToKeyFile) {
		panic("inconsistent TLS certificate and key files - one is insecure and the other is not")
	}

	// If both are insecure, return insecure credentials with no certificates
	if isInsecure(pathToCertFile) && isInsecure(pathToKeyFile) {
		return insecure.NewCredentials(), nil, false, nil
	}

	certificate, err := tls.LoadX509KeyPair(pathToCertFile, pathToKeyFile)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to load TLS certificate or key file: %v", err)
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{certificate},
		MinVersion:   tls.VersionTLS12,
	}), []tls.Certificate{certificate}, true, nil
}
