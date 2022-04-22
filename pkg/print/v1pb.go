package print

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/effective-security/xpki/certutil"
	"github.com/effective-security/xpki/x/print"
	"github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/olekukonko/tablewriter"
)

// ServerVersion prints ServerVersion
func ServerVersion(w io.Writer, r *pb.ServerVersion) {
	fmt.Fprintf(w, "%s (%s)\n", r.Build, r.Runtime)
}

// ServerStatusResponse prints pb.ServerStatusResponse
func ServerStatusResponse(w io.Writer, r *pb.ServerStatusResponse) {
	table := tablewriter.NewWriter(w)
	table.SetBorder(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Append([]string{"Name", r.Status.Name})
	table.Append([]string{"Node", r.Status.Nodename})
	table.Append([]string{"Host", r.Status.Hostname})
	table.Append([]string{"Listen URLs", strings.Join(r.Status.ListenUrls, ",")})
	table.Append([]string{"Version", r.Version.Build})
	table.Append([]string{"Runtime", r.Version.Runtime})

	startedAt := r.Status.StartedAt.AsTime().Local()
	uptime := time.Now().Sub(startedAt) / time.Second * time.Second
	table.Append([]string{"Started", startedAt.Format(time.RFC3339)})
	table.Append([]string{"Uptime", uptime.String()})

	table.Render()
	fmt.Fprintln(w)
}

// CallerStatusResponse prints pb.CallerStatusResponse
func CallerStatusResponse(w io.Writer, r *pb.CallerStatusResponse) {
	table := tablewriter.NewWriter(w)
	table.SetBorder(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Append([]string{"Name", r.Name})
	table.Append([]string{"ID", r.Id})
	table.Append([]string{"Role", r.Role})
	table.Render()
	fmt.Fprintln(w)
}

// Issuers prints list of IssuerInfo
func Issuers(w io.Writer, issuers []*pb.IssuerInfo, withPem bool) {
	now := time.Now()
	for _, ci := range issuers {
		fmt.Fprintf(w, "=========================================================\n")
		certBytes := []byte(ci.Certificate)
		bundle, bundleStatus, err := certutil.VerifyBundleFromPEM(certBytes, []byte(ci.Intermediates), []byte(ci.Root))
		if err != nil {
			fmt.Fprintf(w, "ERROR: unable to create bundle: %s\n", err.Error())

			crt, err := certutil.ParseFromPEM(certBytes)
			if err == nil {
				print.Certificate(w, crt)
			}

			continue
		}

		issuedIn := now.Sub(bundle.Cert.NotBefore.Local()) / time.Minute * time.Minute
		expiresIn := bundle.Cert.NotAfter.Local().Sub(now) / time.Minute * time.Minute

		fmt.Fprintf(w, "Label: %s\n", ci.Label)
		fmt.Fprintf(w, "Profiles: %v\n", ci.Profiles)
		fmt.Fprintf(w, "Subject: %s\n", certutil.NameToString(&bundle.Cert.Subject))
		fmt.Fprintf(w, "  Issuer: %s\n", certutil.NameToString(&bundle.Cert.Issuer))
		fmt.Fprintf(w, "  SKID: %s\n", bundle.SubjectID)
		fmt.Fprintf(w, "  IKID: %s\n", bundle.IssuerID)
		fmt.Fprintf(w, "  Serial: %s\n", bundle.Cert.SerialNumber.String())
		fmt.Fprintf(w, "  Issued: %s (%s ago)\n", bundle.Cert.NotBefore.Local().String(), issuedIn.String())
		fmt.Fprintf(w, "  Expires: %s (in %s)\n", bundle.Cert.NotAfter.Local().String(), expiresIn.String())
		if len(bundleStatus.ExpiringSKIs) > 0 {
			fmt.Fprintf(w, "  Expiring SKI:\n")
			for _, ski := range bundleStatus.ExpiringSKIs {
				fmt.Fprintf(w, "  -- %s\n", ski)
			}
		}
		if len(bundleStatus.Untrusted) > 0 {
			fmt.Fprintf(w, "  Untrusted roots:\n")
			for _, ski := range bundleStatus.Untrusted {
				fmt.Fprintf(w, "  -- %s\n", ski)
			}
		}

		if len(bundle.Cert.CRLDistributionPoints) > 0 {
			fmt.Fprintf(w, "  CRL Distribution Points:\n")
			for _, url := range bundle.Cert.CRLDistributionPoints {
				fmt.Fprintf(w, "  -- %s\n", url)
			}
		}
		if len(bundle.Cert.OCSPServer) > 0 {
			fmt.Fprintf(w, "  OCSP servers:\n")
			for _, url := range bundle.Cert.OCSPServer {
				fmt.Fprintf(w, "  -- %s\n", url)
			}
		}
		if len(bundle.Cert.IssuingCertificateURL) > 0 {
			fmt.Fprintf(w, "  Issuing certificate URL:\n")
			for _, url := range bundle.Cert.IssuingCertificateURL {
				fmt.Fprintf(w, "  -- %s\n", url)
			}
		}

		if len(bundle.Chain) > 1 {
			fmt.Fprintf(w, "Chain:\n")
			cnt := 0
			for _, crt := range bundle.Chain {
				if !bytes.Equal(crt.Raw, bundle.Cert.Raw) {
					cnt++
					fmt.Fprintf(w, "  [%d] %s\n", cnt, certutil.NameToString(&crt.Subject))
					fmt.Fprintf(w, "    SKID: %s\n", certutil.GetSubjectID(crt))
					fmt.Fprintf(w, "    Serial: %s\n", crt.SerialNumber.String())
					fmt.Fprintf(w, "    Issuer: %s\n", certutil.NameToString(&crt.Issuer))
					fmt.Fprintf(w, "    IKID: %s\n", certutil.GetIssuerID(crt))
				}
			}
		} else if bundle.IssuerCert != nil {
			fmt.Fprintf(w, "Issuer: %s\n", certutil.NameToString(&bundle.IssuerCert.Subject))
			fmt.Fprintf(w, "  SKID: %s\n", certutil.GetSubjectID(bundle.IssuerCert))
			fmt.Fprintf(w, "  IKID: %s\n", certutil.GetIssuerID(bundle.IssuerCert))
		}

		if bundle.RootCert != nil &&
			(bundle.IssuerCert == nil || !bytes.Equal(bundle.RootCert.Raw, bundle.IssuerCert.Raw)) {
			fmt.Fprintf(w, "Root: %s\n", certutil.NameToString(&bundle.RootCert.Subject))
			fmt.Fprintf(w, "  SKID: %s\n", certutil.GetSubjectID(bundle.RootCert))
		}

		if withPem {
			fmt.Fprintf(w, "\n%s\n", ci.Certificate)
		}
	}
	fmt.Fprintln(w)
}

// Roots prints list of RootCertificate
func Roots(w io.Writer, roots []*pb.RootCertificate, withPem bool) {
	now := time.Now()

	for cnt, ci := range roots {
		na := ci.NotAfter.AsTime().Local()
		nb := ci.NotBefore.AsTime().Local()
		issuedIn := now.Sub(nb) / time.Minute * time.Minute
		expiresIn := na.Sub(now) / time.Minute * time.Minute

		fmt.Fprintf(w, "==================================== %d ====================================\n", cnt+1)
		fmt.Fprintf(w, "Subject: %s\n", ci.Subject)
		fmt.Fprintf(w, "  ID: %d\n", ci.Id)
		fmt.Fprintf(w, "  SKID: %s\n", ci.Skid)
		fmt.Fprintf(w, "  Thumbprint: %s\n", ci.Sha256)
		fmt.Fprintf(w, "  Trust: %v\n", ci.Trust)
		fmt.Fprintf(w, "  Issued: %s (%s ago)\n", nb.String(), issuedIn.String())
		fmt.Fprintf(w, "  Expires: %s (in %s)\n", na.String(), expiresIn.String())
		if withPem {
			fmt.Fprintf(w, "\n%s\n", ci.Pem)
		}
	}
}

// CertificatesTable prints list of Certificates
func CertificatesTable(w io.Writer, list []*pb.Certificate) {
	table := tablewriter.NewWriter(w)
	table.SetBorder(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader([]string{"Id", "OrgId", "SKID", "Serial", "From", "To", "Subject", "Profile", "Label"})

	for _, c := range list {
		table.Append([]string{
			strconv.FormatUint(c.Id, 10),
			strconv.FormatUint(c.OrgId, 10),
			c.Skid,
			c.SerialNumber,
			c.NotBefore.AsTime().Local().Format(time.RFC3339),
			c.NotAfter.AsTime().Local().Format(time.RFC3339),
			c.Subject,
			c.Profile,
			c.Label,
		})
	}
	table.Render()
	fmt.Fprintln(w)
}

// RevokedCertificatesTable prints list of Revoked Certificates
func RevokedCertificatesTable(w io.Writer, list []*pb.RevokedCertificate) {
	table := tablewriter.NewWriter(w)
	table.SetBorder(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader([]string{"Id", "OrgId", "SKID", "Serial", "From", "To", "Subject", "Profile", "Label", "Revoked", "Reason"})

	for _, r := range list {
		c := r.Certificate
		table.Append([]string{
			strconv.FormatUint(c.Id, 10),
			strconv.FormatUint(c.OrgId, 10),
			c.Skid,
			c.SerialNumber,
			c.NotBefore.AsTime().Local().Format(time.RFC3339),
			c.NotAfter.AsTime().Local().Format(time.RFC3339),
			c.Subject,
			c.Profile,
			c.Label,
			r.RevokedAt.AsTime().Local().Format(time.RFC3339),
			r.Reason.String(),
		})
	}
	table.Render()

	fmt.Fprintln(w)
}

// CrlsTable prints list of CRL
func CrlsTable(w io.Writer, list []*pb.Crl) {
	table := tablewriter.NewWriter(w)
	table.SetBorder(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader([]string{"Id", "IKID", "This Update", "Next Update", "Issuer"})

	for _, c := range list {
		table.Append([]string{
			strconv.FormatUint(c.Id, 10),
			c.Ikid,
			c.ThisUpdate.AsTime().Local().Format(time.RFC3339),
			c.NextUpdate.AsTime().Local().Format(time.RFC3339),
			c.Issuer,
		})
	}
	table.Render()
	fmt.Fprintln(w)
}
