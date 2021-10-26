package commands

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	v2 "github.com/urfave/cli/v2"

	"github.com/cmattoon/dockerenv/pkg/inspector"
)

func TLS() *v2.Command {
	return &v2.Command{
		Name:  "tls",
		Usage: "inspects PEM-encoded TLS certificates",
		Flags: []v2.Flag{},
		Subcommands: []*v2.Command{
			{
				Name:  "verify",
				Usage: "equivalent to 'openssl x509 -noout -text'",
				Flags: []v2.Flag{
					&v2.StringFlag{
						Name:  "cert",
						Usage: "The name of the environment var containing the cert",
					},
					&v2.StringFlag{
						Name:  "key",
						Usage: "The name of the environment var containing the key",
					},
					&v2.StringFlag{
						Name:  "ca-cert",
						Usage: "The name of the environment var containing the CA cert",
					},
					&v2.BoolFlag{
						Name:    "b64decode",
						Aliases: []string{"b64", "d"},
						Usage:   "Apply base64 decoding to the raw value",
					},
				},
				Action: tlsVerifyAction,
			},
		},
	}
}

func tlsVerifyAction(c *v2.Context) (err error) {
	containerId := c.String("container-id")
	if containerId == "" {
		log.Println("Must set container-id")
		return v2.ShowSubcommandHelp(c)
	}
	tlsCertVar := c.String("cert")
	tlsKeyVar := c.String("key")
	tlsCAVar := c.String("ca-cert")

	if tlsCertVar == "" || tlsKeyVar == "" {
		log.Println("Must specify --cert and --key")
		return v2.ShowSubcommandHelp(c)
	}

	i, err := inspector.New()
	if err != nil {
		log.Fatal(err)
	}

	tlsCertPEM, err := i.GetValue(containerId, tlsCertVar)
	if err != nil {
		log.Fatal(err)
	}

	tlsKeyPEM, err := i.GetValue(containerId, tlsKeyVar)
	if err != nil {
		log.Fatal(err)
	}

	tlsCAPEM := ""
	if tlsCAVar != "" {
		tlsCAPEM, err = i.GetValue(containerId, tlsCAVar)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("with CA: %s", tlsCAPEM)
	}

	cert, err := tls.X509KeyPair(fixup("cert", tlsCertPEM), fixup("key", tlsKeyPEM))
	if err != nil {
		log.Fatalf("failed to create keypair from cert+key PEM: %s", err)
	}

	log.Printf("Loaded X509KeyPair with %d certs", len(cert.Certificate))
	for i, crt := range cert.Certificate {
		c, err := x509.ParseCertificate(crt)
		if err != nil {
			log.Println("error parsing certificate %d: %s", i, err)
		}
		// pretty print certificate
		t := fmt.Sprintf("Certificate %d (CA: %v)", i, c.IsCA)
		log.Println(t)
		log.Println(strings.Repeat("=", len(t)))
		log.Printf("\tSubject          : %s", c.Subject)
		log.Printf("\tSubject Key Id   : %x", c.SubjectKeyId)
		log.Println("")
		log.Printf("\tIssuer           : %s", c.Issuer)
		log.Printf("\tAuthority Key Id : %x", c.AuthorityKeyId)
		log.Println("")
		log.Printf("\tNot Before : %s   (%s)", c.NotBefore, TimeElapsed(time.Now(), c.NotBefore, false))
		log.Printf("\tNot After  : %s   (%s)", c.NotAfter, TimeElapsed(time.Now(), c.NotAfter, false))
		log.Println("")
		if c.KeyUsage > 0 {
			log.Printf("\tKey Usage  : %s", prettyKeyUsage(c.KeyUsage))
			log.Println("")
		}

		if len(c.ExtKeyUsage) > 0 {
			log.Printf("\tExtKeyUsages (%d)   : %v", len(c.ExtKeyUsage), c.ExtKeyUsage)
			log.Println("")
		}
		if len(c.UnknownExtKeyUsage) > 0 {
			log.Printf("\tUnkExtKeyUsage (%d) : %v", len(c.UnknownExtKeyUsage), c.UnknownExtKeyUsage)
			log.Println("")
		}
		if len(c.DNSNames) > 0 {
			log.Printf("\tDNSNames (%d)", len(c.DNSNames))
			for _, n := range c.DNSNames {
				log.Printf("\t\t- %s", n)
			}
			log.Println("")
		}
		if len(c.PermittedDNSDomains) > 0 {
			fmt.Printf("\tPermitted DNS Domains (%d)", len(c.PermittedDNSDomains))
			for _, d := range c.PermittedDNSDomains {
				fmt.Printf("\t\t- %s", d)
			}
			log.Println("")
		}
		if len(c.IPAddresses) > 0 {
			log.Printf("\tIPAddresses (%d)", len(c.IPAddresses))
			for _, ip := range c.IPAddresses {
				log.Printf("\t\t- %s", ip)
			}
			log.Println("")
		}

		if len(c.Extensions) > 0 {
			log.Printf("\tExtensions (%d)", len(c.Extensions))
			for _, x := range c.Extensions {

				log.Printf("\t\t- ID: %s", x.Id)
				log.Printf("\t\t  Critical: %v", x.Critical)
				log.Printf("\t\t  Value: %v", x.Value)
			}
			log.Println("")
		}
		if len(c.ExtraExtensions) > 0 {
			log.Printf("\tExtraExtensions (%d)", len(c.ExtraExtensions))
			for _, x := range c.ExtraExtensions {
				log.Printf("\t\t- ID: %s", x.Id)
				log.Printf("\t\t  Critical: %v", x.Critical)
				log.Printf("\t\t  Value: %v", x.Value)
			}
			log.Println("")
		}
		if len(c.UnhandledCriticalExtensions) > 0 {
			log.Printf("\tUnhandledCriticalExtensions (%d)", len(c.UnhandledCriticalExtensions))
			for _, x := range c.UnhandledCriticalExtensions {
				log.Printf("\t\t- ID: %s", x)
			}
			log.Println("")
		}
		log.Println("")

	}
	return nil
}

func fixup(name, pemData string) []byte {
	if strings.Contains(pemData, "\\n") {
		log.Printf("WARNING: detected double escaping in %s", name)
		pemData = strings.ReplaceAll(pemData, "\\n", "\n")
	}

	if strings.HasPrefix(pemData, "\"") && strings.HasSuffix(pemData, "\"") {
		log.Printf("WARNING: removing extra quotes from %s", name)
		pemData = strings.TrimPrefix(pemData, "\"")
		pemData = strings.TrimSuffix(pemData, "\"")
	}
	return []byte(pemData)
}
func s(x float64) string {
	if int(x) == 1 {
		return ""
	}
	return "s"
}

func TimeElapsed(now time.Time, then time.Time, full bool) string {
	var parts []string
	var text string

	year2, month2, day2 := now.Date()
	hour2, minute2, second2 := now.Clock()

	year1, month1, day1 := then.Date()
	hour1, minute1, second1 := then.Clock()

	year := math.Abs(float64(int(year2 - year1)))
	month := math.Abs(float64(int(month2 - month1)))
	day := math.Abs(float64(int(day2 - day1)))
	hour := math.Abs(float64(int(hour2 - hour1)))
	minute := math.Abs(float64(int(minute2 - minute1)))
	second := math.Abs(float64(int(second2 - second1)))

	week := math.Floor(day / 7)

	if year > 0 {
		parts = append(parts, strconv.Itoa(int(year))+" year"+s(year))
	}

	if month > 0 {
		parts = append(parts, strconv.Itoa(int(month))+" month"+s(month))
	}

	if week > 0 {
		parts = append(parts, strconv.Itoa(int(week))+" week"+s(week))
	}

	if day > 0 {
		parts = append(parts, strconv.Itoa(int(day))+" day"+s(day))
	}

	if hour > 0 {
		parts = append(parts, strconv.Itoa(int(hour))+" hour"+s(hour))
	}

	if minute > 0 {
		parts = append(parts, strconv.Itoa(int(minute))+" minute"+s(minute))
	}

	if second > 0 {
		parts = append(parts, strconv.Itoa(int(second))+" second"+s(second))
	}

	if now.After(then) {
		text = " ago"
	} else {
		text = " from now"
	}

	if len(parts) == 0 {
		return "just now"
	}

	if full {
		return strings.Join(parts, ", ") + text
	}
	return parts[0] + text
}

func prettyKeyUsage(u x509.KeyUsage) string {
	switch u {
	case x509.KeyUsageDigitalSignature:
		return "KeyUsageDigitalSignature"
	case x509.KeyUsageContentCommitment:
		return "KeyUsageContentCommitment"
	case x509.KeyUsageKeyEncipherment:
		return "KeyUsageKeyEncipherment"
	case x509.KeyUsageDataEncipherment:
		return "KeyUsageDataEncipherment"
	case x509.KeyUsageKeyAgreement:
		return "KeyUsageKeyAgreement"
	case x509.KeyUsageCertSign:
		return "KeyUsageCertSign"
	case x509.KeyUsageCRLSign:
		return "KeyUsageCRLSign"
	case x509.KeyUsageEncipherOnly:
		return "KeyUsageEncipherOnly"
	case x509.KeyUsageDecipherOnly:
		return "KeyUsageDecipherOnly"
	}
	return ""
}
