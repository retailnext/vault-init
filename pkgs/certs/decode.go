package certs

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

func ParsePEMCertificate(cert string) (*x509.Certificate, error) {
	block, trailing := pem.Decode([]byte(cert))
	if block == nil {
		return nil, fmt.Errorf("PEM decoding failed")
	}
	if len(trailing) > 0 {
		return nil, fmt.Errorf("trailing data (%d bytes) after PEM certificate", len(trailing))
	}
	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("non-certificate PEM block")
	}
	return x509.ParseCertificate(block.Bytes)
}
