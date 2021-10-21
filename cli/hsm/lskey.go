package hsm

import (
	"fmt"
	"time"

	"github.com/go-phorce/dolly/ctl"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/martinisecurity/trusty/cli"
	"github.com/pkg/errors"
)

// LsKeyFlags specifies flags for the Keys action
type LsKeyFlags struct {
	// Token specifies slot token
	Token *string
	// Serial specifies slot serial
	Serial *string
	// Prefix specifies key label prefix
	Prefix *string
}

// Keys shows keys
func Keys(c ctl.Control, p interface{}) error {
	flags := p.(*LsKeyFlags)

	_, def := c.(*cli.Cli).CryptoProv()
	keyProv, ok := def.(cryptoprov.KeyManager)
	if !ok {
		return errors.Errorf("unsupported command for this crypto provider")
	}

	isDefaultSlot := *flags.Serial == "" && *flags.Token == ""
	filterSerial := *flags.Serial
	if filterSerial == "" {
		filterSerial = "--@--"
	}
	filterLabel := *flags.Token
	if filterLabel == "" {
		filterLabel = "--@--"
	}

	printSlot := func(slotID uint, description, label, manufacturer, model, serial string) error {
		if isDefaultSlot || serial == filterSerial || label == filterLabel {
			fmt.Fprintf(c.Writer(), "Slot: %d\n", slotID)
			fmt.Fprintf(c.Writer(), "  Description:  %s\n", description)
			fmt.Fprintf(c.Writer(), "  Token serial: %s\n", serial)
			fmt.Fprintf(c.Writer(), "  Token label:  %s\n", label)

			count := 0
			err := keyProv.EnumKeys(slotID, *flags.Prefix, func(id, label, typ, class, currentVersionID string, creationTime *time.Time) error {
				count++
				fmt.Fprintf(c.Writer(), "[%d]\n", count)
				fmt.Fprintf(c.Writer(), "  Id:    %s\n", id)
				fmt.Fprintf(c.Writer(), "  Label: %s\n", label)
				fmt.Fprintf(c.Writer(), "  Type:  %s\n", typ)
				fmt.Fprintf(c.Writer(), "  Class: %s\n", class)
				fmt.Fprintf(c.Writer(), "  CurrentVersionID:  %s\n", currentVersionID)
				if creationTime != nil {
					fmt.Fprintf(c.Writer(), "  CreationTime: %s\n", creationTime.Format(time.RFC3339))
				}
				return nil
			})
			if err != nil {
				return errors.WithMessagef(err, "failed to list keys on slot %d", slotID)
			}

			if *flags.Prefix != "" && count == 0 {
				fmt.Fprintf(c.Writer(), "no keys found with prefix: %s\n", *flags.Prefix)
			}
		}
		return nil
	}

	return keyProv.EnumTokens(isDefaultSlot, printSlot)
}
