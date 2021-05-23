package print

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/go-phorce/dolly/xpki/certutil"
	"golang.org/x/crypto/ocsp"
)

// Certificates prints list of cert details
func Certificates(w io.Writer, list []*x509.Certificate) {
	for idx, crt := range list {
		fmt.Fprintf(w, "==================================== %d ====================================\n", 1+idx)
		Certificate(w, crt)
	}
}

// Certificate prints cert details
func Certificate(w io.Writer, crt *x509.Certificate) {
	now := time.Now()
	issuedIn := now.Sub(crt.NotBefore.Local()) / time.Minute * time.Minute
	expiresIn := crt.NotAfter.Local().Sub(now) / time.Minute * time.Minute

	fmt.Fprintf(w, "ID: %s\n", certutil.GetSubjectID(crt))
	fmt.Fprintf(w, "Subject: %s\n", certutil.NameToString(&crt.Subject))
	fmt.Fprintf(w, "Serial: %s\n", crt.SerialNumber.String())
	fmt.Fprintf(w, "Issuer: %s\n", certutil.NameToString(&crt.Issuer))
	fmt.Fprintf(w, "Issuer ID: %s\n", certutil.GetIssuerID(crt))
	fmt.Fprintf(w, "Issued: %s (%s ago)\n", crt.NotBefore.Local().String(), issuedIn.String())
	fmt.Fprintf(w, "Expires: %s (in %s)\n", crt.NotAfter.Local().String(), expiresIn.String())
	if len(crt.DNSNames) > 0 {
		fmt.Fprintf(w, "DNS Names:\n")
		for _, n := range crt.DNSNames {
			fmt.Fprintf(w, "  - %s\n", n)
		}
	}
	if len(crt.IPAddresses) > 0 {
		fmt.Fprintf(w, "IP Addresses:\n")
		for _, n := range crt.IPAddresses {
			fmt.Fprintf(w, "  - %s\n", n.String())
		}
	}
	if len(crt.URIs) > 0 {
		fmt.Fprintf(w, "URIs:\n")
		for _, n := range crt.URIs {
			fmt.Fprintf(w, "  - %s\n", n.String())
		}
	}
	if len(crt.EmailAddresses) > 0 {
		fmt.Fprintf(w, "Emails:\n")
		for _, n := range crt.EmailAddresses {
			fmt.Fprintf(w, "  - %s\n", n)
		}
	}
	if len(crt.CRLDistributionPoints) > 0 {
		fmt.Fprintf(w, "CRL Distribution Points:\n")
		for _, n := range crt.CRLDistributionPoints {
			fmt.Fprintf(w, "  - %s\n", n)
		}
	}
	if len(crt.OCSPServer) > 0 {
		fmt.Fprintf(w, "OCSP Servers:\n")
		for _, n := range crt.OCSPServer {
			fmt.Fprintf(w, "  - %s\n", n)
		}
	}
	if len(crt.IssuingCertificateURL) > 0 {
		fmt.Fprintf(w, "Issuing Certificates:\n")
		for _, n := range crt.IssuingCertificateURL {
			fmt.Fprintf(w, "  - %s\n", n)
		}
	}
	if crt.IsCA {
		fmt.Fprintf(w, "CA: true\n")
		fmt.Fprintf(w, "  Basic Constraints Valid: %t\n", crt.BasicConstraintsValid)
		fmt.Fprintf(w, "  Max Path: %d\n", crt.MaxPathLen)
	}
}

// CertificateRequest prints cert request details
func CertificateRequest(w io.Writer, crt *x509.CertificateRequest) {
	fmt.Fprintf(w, "Subject: %s\n", certutil.NameToString(&crt.Subject))
	if len(crt.DNSNames) > 0 {
		fmt.Fprintf(w, "DNS Names:\n")
		for _, n := range crt.DNSNames {
			fmt.Fprintf(w, "  - %s\n", n)
		}
	}
	if len(crt.IPAddresses) > 0 {
		fmt.Fprintf(w, "IP Addresses:\n")
		for _, n := range crt.IPAddresses {
			fmt.Fprintf(w, "  - %s\n", n.String())
		}
	}
	if len(crt.URIs) > 0 {
		fmt.Fprintf(w, "URIs:\n")
		for _, n := range crt.URIs {
			fmt.Fprintf(w, "  - %s\n", n.String())
		}
	}
	if len(crt.EmailAddresses) > 0 {
		fmt.Fprintf(w, "Emails:\n")
		for _, n := range crt.EmailAddresses {
			fmt.Fprintf(w, "  - %s\n", n)
		}
	}
	if len(crt.Extensions) > 0 {
		fmt.Fprintf(w, "Extensions:\n")
		for _, n := range crt.Extensions {
			fmt.Fprintf(w, "  - %s\n", n.Id.String())
		}
	}
}

// CertificateList prints CRL details
func CertificateList(w io.Writer, crl *pkix.CertificateList) {
	now := time.Now()
	issuedIn := now.Sub(crl.TBSCertList.ThisUpdate) / time.Minute * time.Minute
	expiresIn := crl.TBSCertList.NextUpdate.Sub(now) / time.Minute * time.Minute

	fmt.Fprintf(w, "Version: %d\n", crl.TBSCertList.Version)
	fmt.Fprintf(w, "Issuer: %s\n", crl.TBSCertList.Issuer.String())
	fmt.Fprintf(w, "Issued: %s (%s ago)\n", crl.TBSCertList.ThisUpdate.Local().String(), issuedIn.String())
	fmt.Fprintf(w, "Expires: %s (in %s)\n", crl.TBSCertList.NextUpdate.Local().String(), expiresIn.String())

	if len(crl.TBSCertList.RevokedCertificates) > 0 {
		fmt.Fprintf(w, "Revoked:\n")
		for _, r := range crl.TBSCertList.RevokedCertificates {
			fmt.Fprintf(w, "  - %s | %s\n",
				r.SerialNumber.String(),
				r.RevocationTime.Local().Format(time.RFC3339))
		}
	}
}

var ocspStatusCode = map[int]string{
	ocsp.Good:    "good",
	ocsp.Revoked: "revoked",
	ocsp.Unknown: "unknown",
}

// OCSPResponse prints OCSP response details
func OCSPResponse(w io.Writer, res *ocsp.Response) {
	now := time.Now()
	issuedIn := now.Sub(res.ProducedAt) / time.Minute * time.Minute
	updatedIn := now.Sub(res.ThisUpdate) / time.Minute * time.Minute
	expiresIn := res.NextUpdate.Sub(now) / time.Minute * time.Minute

	fmt.Fprintf(w, "Serial: %s\n", res.SerialNumber.String())
	fmt.Fprintf(w, "Issued: %s (%s ago)\n", res.ProducedAt.Local().String(), issuedIn.String())
	fmt.Fprintf(w, "Updated: %s (%s ago)\n", res.ThisUpdate.Local().String(), updatedIn.String())
	fmt.Fprintf(w, "Expires: %s (in %s)\n", res.NextUpdate.Local().String(), expiresIn.String())
	fmt.Fprintf(w, "Status: %s\n", ocspStatusCode[res.Status])
	if res.Status == ocsp.Revoked {
		fmt.Fprintf(w, "Revocation reason: %d\n", res.RevocationReason)
	}
}

// CSRandCert outputs a cert, key and csr
func CSRandCert(w io.Writer, key, csrBytes, cert []byte) {
	out := map[string]string{}
	if cert != nil {
		out["cert"] = string(cert)
	}

	if key != nil {
		out["key"] = string(key)
	}

	if csrBytes != nil {
		out["csr"] = string(csrBytes)
	}

	jsonOut, _ := json.Marshal(out)
	fmt.Fprintln(w, string(jsonOut))
}
