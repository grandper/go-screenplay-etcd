package ability

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/grandper/go-screenplay-etcd/etcd"
)

// WithTheCertificateAuthority makes the ability trust the servers whose
// certificate is signed by the authority of the given PEM file; it reads
// "with the certificate authority ca.crt". Without it, the authorities of the
// system are trusted. The connection is then secured with TLS.
func (a *UseETCDAbility) WithTheCertificateAuthority(caFile string) *UseETCDAbility {
	if a == nil {
		return nil
	}

	a.connecting.Lock()
	defer a.connecting.Unlock()

	a.caFile = caFile

	return a
}

// WithTheClientCertificate gives the ability the certificate it presents to
// the server, for a cluster that authenticates its clients with mutual TLS;
// it reads "with the client certificate client.crt and the key client.key"
// once closed with AndTheKey. The connection is then secured with TLS.
func (a *UseETCDAbility) WithTheClientCertificate(certFile string) *ClientCertificateBuilder {
	return &ClientCertificateBuilder{ability: a, certFile: certFile}
}

// ClientCertificateBuilder holds the certificate given to
// WithTheClientCertificate and waits for its key.
type ClientCertificateBuilder struct {
	ability  *UseETCDAbility
	certFile string
}

// AndTheKey gives the ability the private key of the client certificate, as
// a PEM file.
func (b *ClientCertificateBuilder) AndTheKey(keyFile string) *UseETCDAbility {
	if b.ability == nil {
		return nil
	}

	b.ability.connecting.Lock()
	defer b.ability.connecting.Unlock()

	b.ability.certFile = b.certFile
	b.ability.keyFile = keyFile

	return b.ability
}

// tlsConfig builds the TLS configuration of the client from the certificate
// authority and the client certificate given to the ability, or nil when none
// was given: the connection is then in plain text.
func (a *UseETCDAbility) tlsConfig() (*tls.Config, error) {
	if a.caFile == "" && a.certFile == "" && a.keyFile == "" {
		return nil, nil //nolint:nilnil // no configuration means a plain-text connection
	}

	config := &tls.Config{MinVersion: tls.VersionTLS12}

	if a.caFile != "" {
		authorities, err := os.ReadFile(a.caFile)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", etcd.ErrFailedToLoadTheCertificates, err)
		}

		config.RootCAs = x509.NewCertPool()
		if !config.RootCAs.AppendCertsFromPEM(authorities) {
			return nil, fmt.Errorf("%w: %q holds no certificate", etcd.ErrFailedToLoadTheCertificates, a.caFile)
		}
	}

	if a.certFile != "" || a.keyFile != "" {
		certificate, err := tls.LoadX509KeyPair(a.certFile, a.keyFile)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", etcd.ErrFailedToLoadTheCertificates, err)
		}

		config.Certificates = []tls.Certificate{certificate}
	}

	return config, nil
}
