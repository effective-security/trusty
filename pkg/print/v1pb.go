package print

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/effective-security/trusty/api/v1/pb"
	"github.com/effective-security/xpki/certutil"
	"github.com/effective-security/xpki/jwt"
	"github.com/effective-security/xpki/x/print"
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
	table.Append([]string{"Listen URLs", strings.Join(r.Status.ListenURLs, ",")})
	table.Append([]string{"Version", r.Version.Build})
	table.Append([]string{"Runtime", r.Version.Runtime})
	table.Append([]string{"Started", r.Status.StartedAt})

	table.Render()
	fmt.Fprintln(w)
}

// CallerStatusResponse prints pb.CallerStatusResponse
func CallerStatusResponse(w io.Writer, r *pb.CallerStatusResponse) {
	table := tablewriter.NewWriter(w)
	table.SetBorder(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)
	table.Append([]string{"Subject", r.Subject})
	table.Append([]string{"Role", r.Role})

	now := time.Now().Local()
	var claims jwt.MapClaims
	_ = json.Unmarshal(r.Claims, &claims)
	for k := range claims {
		var val string
		if dateClaims[k] {
			ptim := claims.Time(k)
			if ptim != nil {
				tim := ptim.Local()
				if k == "exp" {
					inMins := tim.Sub(now) / time.Minute * time.Minute
					val = fmt.Sprintf("%s, in %s", tim.Format(time.RFC3339), inMins.String())
				} else {
					ago := now.Sub(tim) / time.Minute * time.Minute
					val = fmt.Sprintf("%s, %s ago", tim.Format(time.RFC3339), ago.String())
				}
			} else {
				val = claims.String(k)
			}
		} else {
			val = claims.String(k)
		}

		table.Append([]string{"claim:" + k, val})
	}

	table.Render()
	fmt.Fprintln(w)
}

var dateClaims = map[string]bool{
	"iat": true,
	"nbf": true,
	"exp": true,
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
				print.Certificate(w, crt, false)
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
	for cnt, ci := range roots {
		fmt.Fprintf(w, "==================================== %d ====================================\n", cnt+1)
		fmt.Fprintf(w, "Subject: %s\n", ci.Subject)
		fmt.Fprintf(w, "  ID: %d\n", ci.ID)
		fmt.Fprintf(w, "  SKID: %s\n", ci.SKID)
		fmt.Fprintf(w, "  Thumbprint: %s\n", ci.Sha256)
		fmt.Fprintf(w, "  Trust: %v\n", ci.Trust)
		fmt.Fprintf(w, "  Issued: %s\n", ci.NotAfter)
		fmt.Fprintf(w, "  Expires: %s\n", ci.NotBefore)
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
	table.SetHeader([]string{"Id", "OrgID", "SKID", "Serial", "From", "To", "Subject", "Profile", "Label"})

	for _, c := range list {
		table.Append([]string{
			strconv.FormatUint(c.ID, 10),
			strconv.FormatUint(c.OrgID, 10),
			c.SKID,
			c.SerialNumber,
			c.NotBefore,
			c.NotAfter,
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
	table.SetHeader([]string{"Id", "OrgID", "SKID", "Serial", "From", "To", "Subject", "Profile", "Label", "Revoked", "Reason"})

	for _, r := range list {
		c := r.Certificate
		table.Append([]string{
			strconv.FormatUint(c.ID, 10),
			strconv.FormatUint(c.OrgID, 10),
			c.SKID,
			c.SerialNumber,
			c.NotBefore,
			c.NotAfter,
			c.Subject,
			c.Profile,
			c.Label,
			r.RevokedAt,
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
			strconv.FormatUint(c.ID, 10),
			c.IKID,
			c.ThisUpdate,
			c.NextUpdate,
			c.Issuer,
		})
	}
	table.Render()
	fmt.Fprintln(w)
}

// RevokedCertificate prints RevokedCertificate
func RevokedCertificate(w io.Writer, ci *pb.RevokedCertificate, withPem bool) {
	fmt.Fprintf(w, "Revoked: %s\n", ci.RevokedAt)
	fmt.Fprintf(w, "  Reason: %s\n", ci.Reason)
	Certificate(w, ci.Certificate, withPem)
}

// Certificate prints Certificate
func Certificate(w io.Writer, ci *pb.Certificate, withPem bool) {
	fmt.Fprintf(w, "Subject: %s\n", ci.Subject)
	fmt.Fprintf(w, "  Issuer: %s\n", ci.Issuer)
	fmt.Fprintf(w, "  ID: %d\n", ci.ID)
	fmt.Fprintf(w, "  SKID: %s\n", ci.SKID)
	fmt.Fprintf(w, "  SN: %s\n", ci.SerialNumber)
	fmt.Fprintf(w, "  Thumbprint: %s\n", ci.Sha256)
	fmt.Fprintf(w, "  Issued: %s\n", ci.NotAfter)
	fmt.Fprintf(w, "  Expires: %s\n", ci.NotBefore)
	fmt.Fprintf(w, "  Profile: %s\n", ci.Profile)
	if len(ci.Locations) > 0 {
		fmt.Fprintf(w, "  Locations:\n")
		for _, v := range ci.Locations {
			fmt.Fprintf(w, "    %s\n", v)
		}

	}
	if len(ci.Metadata) > 0 {
		fmt.Fprintf(w, "  Metadata:\n")
		for k, v := range ci.Metadata {
			fmt.Fprintf(w, "    %s: %s\n", k, v)
		}
	}
	if withPem {
		fmt.Fprintf(w, "\n%s\n", ci.Pem)
	}
}
