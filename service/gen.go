package service

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/aacfactory/afssl"
)

func Gen(output string) (err error) {
	if output == "" {
		output = "."
	}
	if !filepath.IsAbs(output) {
		abs, absErr := filepath.Abs(output)
		if absErr != nil {
			err = errors.Join(errors.New("generate ca keypair failed"), errors.New("could not resolve output path"), absErr)
			return
		}
		output = abs
	}
	output = filepath.ToSlash(output)

	config := afssl.CertificateConfig{
		Issuer: &afssl.CertificatePkixName{
			Country:            "CN",
			Province:           "",
			Locality:           "",
			Organization:       "CNXGM",
			OrganizationalUnit: "",
			StreetAddress:      "",
			PostalCode:         "",
			SerialNumber:       "",
			CommonName:         "poweroffs",
		},
		Subject:  nil,
		IPs:      nil,
		Emails:   nil,
		DNSNames: nil,
	}
	caPEM, caKeyPEM, caErr := afssl.GenerateCertificate(config, afssl.CA(), afssl.WithExpirationDays(365*5))
	if caErr != nil {
		err = errors.Join(errors.New("generate ca keypair failed"), caErr)
		return
	}
	caFilename := filepath.Join(output, "ca.crt")
	keyFilename := filepath.Join(output, "ca.key")
	if err = os.WriteFile(caFilename, caPEM, 0600); err != nil {
		err = errors.Join(errors.New("generate ca keypair failed"), errors.New("could not write ca keypair"), caErr)
		return
	}
	if err = os.WriteFile(keyFilename, caKeyPEM, 0600); err != nil {
		_ = os.Remove(caFilename)
		err = errors.Join(errors.New("generate ca keypair failed"), errors.New("could not write ca keypair"), caErr)
		return
	}
	return
}
