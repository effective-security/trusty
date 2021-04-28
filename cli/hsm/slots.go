package hsm

import (
	"fmt"

	"github.com/ekspand/trusty/cli"
	"github.com/go-phorce/dolly/ctl"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/juju/errors"
)

// Slots shows hsm slots
func Slots(c ctl.Control, _ interface{}) error {
	_, defprov := c.(*cli.Cli).CryptoProv()
	keyProv, ok := defprov.(cryptoprov.KeyManager)
	if !ok {
		return errors.Errorf("unsupported command for this crypto provider")
	}

	err := keyProv.EnumTokens(false, func(slotID uint, description, label, manufacturer, model, serial string) error {
		fmt.Fprintf(c.Writer(), "Slot: %d\n", slotID)
		fmt.Fprintf(c.Writer(), "  Description:  %s\n", description)
		fmt.Fprintf(c.Writer(), "  Token serial: %s\n", serial)
		fmt.Fprintf(c.Writer(), "  Token label:  %s\n", label)
		return nil
	})
	if err != nil {
		return errors.Annotate(err, "unable to list slots")
	}

	return nil
}
