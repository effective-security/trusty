package hsm

import (
	"fmt"
	"time"

	"github.com/go-phorce/dolly/ctl"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/cli"
)

// KeyInfoFlags specifies flags for the key info action
type KeyInfoFlags struct {
	// Token specifies slot token
	Token *string
	// Serial specifies slot serial
	Serial *string
	// ID specifies key Id
	ID *string
	// Public specifies if public key should be listed
	Public *bool
}

// KeyInfo retrieves info about a key
func KeyInfo(c ctl.Control, p interface{}) error {
	flags := p.(*KeyInfoFlags)

	_, defprov := c.(*cli.Cli).CryptoProv()
	keyProv, ok := defprov.(cryptoprov.KeyManager)
	if !ok {
		return errors.Errorf("unsupported command for this crypto provider")
	}

	filterSerial := *flags.Serial
	isDefaultSlot := filterSerial == ""

	if isDefaultSlot {
		filterSerial = "--@--"
	}

	slotCount := 0
	printSlot := func(slotID uint, description, label, manufacturer, model, serial string) error {
		if isDefaultSlot || serial == filterSerial {
			slotCount++

			fmt.Fprintf(c.Writer(), "Slot: %d\n", slotID)
			fmt.Fprintf(c.Writer(), "  Description:  %s\n", description)
			fmt.Fprintf(c.Writer(), "  Token serial: %s\n", serial)

			count := 0
			err := keyProv.KeyInfo(slotID, *flags.ID, *flags.Public, func(id, label, typ, class, currentVersionID, pubKey string, creationTime *time.Time) error {
				count++
				fmt.Fprintf(c.Writer(), "[%d]\n", count)
				fmt.Fprintf(c.Writer(), "  Id:    %s\n", id)
				fmt.Fprintf(c.Writer(), "  Label: %s\n", label)
				fmt.Fprintf(c.Writer(), "  Type:  %s\n", typ)
				fmt.Fprintf(c.Writer(), "  Class: %s\n", class)
				fmt.Fprintf(c.Writer(), "  CurrentVersionID: %s\n", currentVersionID)
				fmt.Fprintf(c.Writer(), "  Public key: \n%s\n", pubKey)
				if creationTime != nil {
					fmt.Fprintf(c.Writer(), "  CreationTime: %s\n", creationTime.Format(time.RFC3339))
				}
				return nil
			})
			if err != nil {
				fmt.Fprintf(c.Writer(), "failed to get key info on slot %d, keyID %s: %v\n", slotID, *flags.ID, err)
				return nil
			}
		}
		return nil
	}

	keyProv.EnumTokens(isDefaultSlot, printSlot)

	if slotCount == 0 {
		fmt.Fprintf(c.Writer(), "no slots found with serial: %s\n", filterSerial)
	}

	return nil
}
