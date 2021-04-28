package main

import (
	"fmt"
	"io"
	"os"

	"github.com/ekspand/trusty/cli"
	"github.com/ekspand/trusty/cli/csr"
	"github.com/ekspand/trusty/cli/hsm"
	"github.com/ekspand/trusty/internal/version"
	"github.com/go-phorce/dolly/ctl"
	"github.com/go-phorce/dolly/xlog"
)

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/cmd", "trusty-tool")

const (
	rcError   = 1
	rcSuccess = 0
)

func main() {
	// Logs are set to os.Stderr, while output to os.Stdout
	rc := realMain(os.Args, os.Stdout, os.Stderr)
	os.Exit(int(rc))
}

func realMain(args []string, out io.Writer, errout io.Writer) ctl.ReturnCode {
	formatter := xlog.NewColorFormatter(errout, true)
	xlog.SetFormatter(formatter)

	app := ctl.NewApplication("trusty-tool", "A command-line utility for issuing offline certificates.").
		UsageWriter(out).
		Writer(out).
		ErrorWriter(errout).
		Version(fmt.Sprintf("trusty-tool %v", version.Current().String()))

	cli := cli.New(
		&ctl.ControlDefinition{
			App:    app,
			Output: out,
		},
		cli.WithHsmCfg(),
	)
	defer cli.Close()

	// csr self-sign|create|sign|gencert
	cmdCSR := app.Command("csr", "CSR commands").
		PreAction(cli.PopulateControl).
		PreAction(cli.EnsureCryptoProvider)
		/*
			createRootFlags := new(csr.RootFlags)
			cmdCreateRoot := cmdCSR.Command("self-sign", "generates key and creates self-signed certificate").
				Action(cli.RegisterAction(csr.Root, createRootFlags))
			createRootFlags.CsrProfile = cmdCreateRoot.Flag("csr-profile", "CSR profile file").Required().String()
			createRootFlags.KeyLabel = cmdCreateRoot.Flag("key-label", "label for generated key").String()
			createRootFlags.Output = cmdCreateRoot.Flag("out", "specifies the optional prefix for output files").String()
		*/

	createCSRFlags := new(csr.CreateFlags)
	cmdCreateCSR := cmdCSR.Command("create", "generates key and creates certificate request").
		Action(cli.RegisterAction(csr.Create, createCSRFlags))
	createCSRFlags.CsrProfile = cmdCreateCSR.Flag("csr-profile", "CSR profile file").Required().String()
	createCSRFlags.KeyLabel = cmdCreateCSR.Flag("key-label", "label for generated key").String()
	createCSRFlags.Output = cmdCreateCSR.Flag("out", "specifies the optional prefix for output files").String()

	signCSRFlags := new(csr.SignFlags)
	cmdSignCSR := cmdCSR.Command("sign", "signs certificate request with provided CA key").
		Action(cli.RegisterAction(csr.Sign, signCSRFlags))
	signCSRFlags.Csr = cmdSignCSR.Flag("csr", "CSR file to be signed").Required().String()
	signCSRFlags.CAConfig = cmdSignCSR.Flag("ca-config", "CA configuration file").Required().String()
	signCSRFlags.CACert = cmdSignCSR.Flag("ca-cert", "CA certificate").Required().String()
	signCSRFlags.CAKey = cmdSignCSR.Flag("ca-key", "CA key").Required().String()
	signCSRFlags.Profile = cmdSignCSR.Flag("profile", "certificate profile").Required().String()
	signCSRFlags.SAN = cmdSignCSR.Flag("SAN", "coma separated list of SAN to be added to certificate").String()
	signCSRFlags.Output = cmdSignCSR.Flag("out", "specifies the optional prefix for output files").String()

	genCertFlags := new(csr.GenCertFlags)
	cmdGenCertCSR := cmdCSR.Command("gencert", "creates certificate with provided CA key").
		Action(cli.RegisterAction(csr.GenCert, genCertFlags))
	genCertFlags.SelfSign = cmdGenCertCSR.Flag("self-sign", "enables to create a self-signed certificate").Bool()
	genCertFlags.CsrProfile = cmdGenCertCSR.Flag("csr-profile", "CSR file to be signed").Required().String()
	genCertFlags.CAConfig = cmdGenCertCSR.Flag("ca-config", "CA configuration file").Required().String()
	genCertFlags.CACert = cmdGenCertCSR.Flag("ca-cert", "CA certificate").String()
	genCertFlags.CAKey = cmdGenCertCSR.Flag("ca-key", "CA key").String()
	genCertFlags.Profile = cmdGenCertCSR.Flag("profile", "certificate profile").Required().String()
	genCertFlags.KeyLabel = cmdGenCertCSR.Flag("key-label", "label for generated key").String()
	genCertFlags.SAN = cmdGenCertCSR.Flag("SAN", "coma separated list of SAN to be added to certificate").String()
	genCertFlags.Output = cmdGenCertCSR.Flag("out", "specifies the optional prefix for output files").String()

	// hsm slots|lskey|rmkey|genkey
	cmdHsm := app.Command("hsm", "Perform HSM operations").
		PreAction(cli.PopulateControl).
		PreAction(cli.EnsureCryptoProvider)

	cmdHsm.Command("slots", "Show available slots list").Action(cli.RegisterAction(hsm.Slots, nil))

	hsmLsKeyFlags := new(hsm.LsKeyFlags)
	cmdHsmKeys := cmdHsm.Command("lskey", "Show keys list").Action(cli.RegisterAction(hsm.Keys, hsmLsKeyFlags))
	hsmLsKeyFlags.Token = cmdHsmKeys.Flag("token", "slot token").String()
	hsmLsKeyFlags.Serial = cmdHsmKeys.Flag("serial", "slot serial").String()
	hsmLsKeyFlags.Prefix = cmdHsmKeys.Flag("prefix", "key label prefix").String()

	hsmRmKeyFlags := new(hsm.RmKeyFlags)
	cmdRmKey := cmdHsm.Command("rmkey", "Destroy key").Action(cli.RegisterAction(hsm.RmKey, hsmRmKeyFlags))
	hsmRmKeyFlags.Token = cmdRmKey.Flag("token", "slot token").String()
	hsmRmKeyFlags.Serial = cmdRmKey.Flag("serial", "slot serial").String()
	hsmRmKeyFlags.ID = cmdRmKey.Flag("id", "key ID").String()
	hsmRmKeyFlags.Prefix = cmdRmKey.Flag("prefix", "remove keys based on the specified label prefix").String()
	hsmRmKeyFlags.Force = cmdRmKey.Flag("force", "do not ask for confirmation to remove keys").Bool()

	hsmKeyInfoFlags := new(hsm.KeyInfoFlags)
	cmdKeyInfo := cmdHsm.Command("keyinfo", "Get key info").Action(cli.RegisterAction(hsm.KeyInfo, hsmKeyInfoFlags))
	hsmKeyInfoFlags.Token = cmdKeyInfo.Flag("token", "slot token").String()
	hsmKeyInfoFlags.Serial = cmdKeyInfo.Flag("serial", "slot serial").String()
	hsmKeyInfoFlags.ID = cmdKeyInfo.Flag("id", "key ID").Required().String()
	hsmKeyInfoFlags.Public = cmdKeyInfo.Flag("public", "include public key").Bool()

	hsmGenKeyFlags := new(hsm.GenKeyFlags)
	cmdHsmGenKey := cmdHsm.Command("genkey", "Generate key").Action(cli.RegisterAction(hsm.GenKey, hsmGenKeyFlags))
	hsmGenKeyFlags.Purpose = cmdHsmGenKey.Flag("purpose", "Key purpose: signing|encryption").Required().String()
	hsmGenKeyFlags.Algo = cmdHsmGenKey.Flag("alg", "Key algorithm: ECDSA|RSA").Required().String()
	hsmGenKeyFlags.Size = cmdHsmGenKey.Flag("size", "Key size in bits").Required().Int()
	hsmGenKeyFlags.Label = cmdHsmGenKey.Flag("label", "Label for generated key").String()
	hsmGenKeyFlags.Output = cmdHsmGenKey.Flag("output", "Optional output file name").String()
	hsmGenKeyFlags.Force = cmdHsmGenKey.Flag("force", "Override output file if exists").Bool()

	cli.Parse(args)
	return cli.ReturnCode()
}
