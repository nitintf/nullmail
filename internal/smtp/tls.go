package smtp

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log/slog"
	"math/big"
	"net"
	"os"
	"time"
)

// generateSelfSignedCert creates a self-signed certificate for development
func generateSelfSignedCert() (tls.Certificate, error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Nullmail Development"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:     []string{"localhost", "temp-smtp.local", "nullmail.local", "smtp.nullmail.nitin.sh", "nullmail.nitin.sh"},
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Convert to PEM format
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Create TLS certificate
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, err
	}

	slog.Info("Generated self-signed TLS certificate for development")
	return tlsCert, nil
}

// loadOrGenerateTLSConfig loads existing certs or generates new ones
func loadOrGenerateTLSConfig() *tls.Config {
	// Try to load existing certificate files
	if _, err := os.Stat("server.crt"); err == nil {
		if _, err := os.Stat("server.key"); err == nil {
			cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
			if err == nil {
				slog.Info("Loaded existing TLS certificate")
				return &tls.Config{
					Certificates: []tls.Certificate{cert},
					ServerName:   "temp-smtp.local",
				}
			}
		}
	}

	// Generate self-signed certificate for development
	cert, err := generateSelfSignedCert()
	if err != nil {
		slog.Error("Failed to generate TLS certificate", "error", err)
		return nil
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ServerName:   "temp-smtp.local",
	}
}