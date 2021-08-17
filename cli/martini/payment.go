package martini

import (
	"bytes"
	"html/template"

	"github.com/go-phorce/dolly/ctl"
	"github.com/juju/errors"
	"github.com/pkg/browser"
)

// PayOrgFlags defines flags for PayOrg command
type PayOrgFlags struct {
	StripeKey    *string
	ClientSecret *string
	NoBrowser    *bool
}

// PayOrg pays for organization
func PayOrg(c ctl.Control, p interface{}) error {
	flags := p.(*PayOrgFlags)

	tmpl, err := template.New("payment").Parse(paymentTemplate)
	if err != nil {
		return errors.Annotatef(err, "failed to parse template")
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]string{
		"ClientSecret": *flags.ClientSecret,
		"StripeKey":    *flags.StripeKey,
	}); err != nil {
		return errors.Annotatef(err, "failed to render payment template")
	}

	if flags.NoBrowser == nil || !*flags.NoBrowser {
		browser.OpenReader(bytes.NewReader(buf.Bytes()))
	}

	return nil
}

const paymentTemplate = `
<!DOCTYPE html>
<html>
  <head>
    <title>Subscribe</title>
    <meta name="viewport" content="width=device-width, initial-scale=1" />

    <link href="css/base.css" rel="stylesheet" />
    <script src="https://js.stripe.com/v3/"></script>
    <script src="subscribe.js" defer></script>
  </head>
  <body>
    <main>
      <h1>Subscribe</h1>

      <p>
        Try the successful test card: <span>4242424242424242</span>.
      </p>

      <p>
        Try the test card that requires SCA: <span>4000002500003155</span>.
      </p>

      <p>
        Use any <i>future</i> expiry date, CVC, and 5 digit postal code.
      </p>

      <hr />

      <form id="subscribe-form">
        <label>
          Full name
          <input type="text" id="name" value="Hayk Baluyan" />
        </label>

        <div id="card-element">
          <!-- the card element will be mounted here -->
        </div>

        <button type="submit">
          Subscribe
        </button>

        <div id="messages"></div>
      </form>
    </main>
  </body>

  <script>
 // helper method for displaying a status message.
const setMessage = (message) => {
	const messageDiv = document.querySelector('#messages');
	messageDiv.innerHTML += "<br>" + message;
  }
  
  // Initialize an instance of Stripe
  const stripe = Stripe("{{.StripeKey}}");
  
  // Create and mount the single line card element
  const elements = stripe.elements();
  const cardElement = elements.create('card');
  cardElement.mount('#card-element');
  
  // Payment info collection and confirmation
  // When the submit button is pressed, attempt to confirm the payment intent
  // with the information input into the card element form.
  // - handle payment errors by displaying an alert. The customer can update
  //   the payment information and try again
  // - Stripe Elements automatically handles next actions like 3DSecure that are required for SCA
  // - Complete the subscription flow when the payment succeeds
  const form = document.querySelector('#subscribe-form');
  form.addEventListener('submit', async (e) => {
	e.preventDefault();
	
	// Create payment method and confirm payment intent.
	stripe.confirmCardPayment({{.ClientSecret}}, {
	  payment_method: {
		card: cardElement,
	  }
	}).then((result) => {
	  if(result.error) {
		setMessage(` + "`Payment failed: ${result.error.message}`" + `);
	  } else {
		setMessage('Success! You can redirect to another page.');
	  }
	});
  });

  </script>
</html>
`
