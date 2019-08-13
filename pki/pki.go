package main

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"log"
	"os"
	"time"

	"github.com/google/easypki/pkg/certificate"
	"gopkg.in/google/easypki.v1/pkg/easypki"
	"gopkg.in/google/easypki.v1/pkg/store"
)

// PKIConfig config for pki
type PKIConfig struct {
	Root               string
	Organization       string
	OrganizationalUnit string
	Country            string
	Locality           string
	Province           string
}

func main() {
	conf := &PKIConfig{
		Root:               "test",
		Organization:       "ScreepsPlus",
		OrganizationalUnit: "IT",
		Country:            "US",
		Locality:           "Florence",
		Province:           "AL",
	}
	os.Mkdir(conf.Root, os.ModeDir|os.ModePerm)
	pki := &easypki.EasyPKI{Store: &store.Local{Root: conf.Root}}

	var signer *certificate.Bundle
	// signer, err := pki.GetCA("root")
	signer, err := pki.GetCA("IntA")
	if err != nil {
		log.Fatal(err)
	}

	// req := createCA(conf)
	req := createUser(conf, "admin", []string{})
	if err := pki.Sign(signer, req); err != nil {
		log.Fatal(err)
	}
}

func createCA(conf *PKIConfig) *easypki.Request {
	subject := pkix.Name{
		CommonName:         "RootCA",
		Organization:       []string{conf.Organization},
		OrganizationalUnit: []string{conf.OrganizationalUnit},
		Country:            []string{conf.Country},
		Locality:           []string{conf.Locality},
		Province:           []string{conf.Province},
	}
	template := &x509.Certificate{
		Subject:    subject,
		NotAfter:   time.Now().AddDate(20, 0, 0),
		MaxPathLen: 2,
		IsCA:       true,
	}
	return &easypki.Request{
		Name:                "root",
		Template:            template,
		IsClientCertificate: false,
		PrivateKeySize:      4096,
	}
}

func createInt(conf *PKIConfig, name string) *easypki.Request {
	subject := pkix.Name{
		CommonName:         name,
		Organization:       []string{conf.Organization},
		OrganizationalUnit: []string{conf.OrganizationalUnit},
		Country:            []string{conf.Country},
		Locality:           []string{conf.Locality},
		Province:           []string{conf.Province},
	}
	template := &x509.Certificate{
		Subject:    subject,
		NotAfter:   time.Now().AddDate(10, 0, 0),
		MaxPathLen: 1,
		IsCA:       true,
	}
	return &easypki.Request{
		Name:                name,
		Template:            template,
		IsClientCertificate: false,
		PrivateKeySize:      4096,
	}
}

func createUser(conf *PKIConfig, name string, orgs []string) *easypki.Request {
	hasName := false
	for _, v := range orgs {
		if v == name {
			hasName = true
			break
		}
	}
	if !hasName {
		orgs = append(orgs, name)
	}
	subject := pkix.Name{
		CommonName:         name,
		Organization:       []string{conf.Organization},
		OrganizationalUnit: []string{conf.OrganizationalUnit},
		Country:            []string{conf.Country},
		Locality:           []string{conf.Locality},
		Province:           []string{conf.Province},
	}
	template := &x509.Certificate{
		Subject:    subject,
		NotAfter:   time.Now().AddDate(0, 0, 7),
		MaxPathLen: 0,
		IsCA:       false,
		DNSNames:   orgs,
	}
	return &easypki.Request{
		Name:                name,
		Template:            template,
		IsClientCertificate: true,
		PrivateKeySize:      4096,
	}
}
