// helper method for displaying a status message.
const setMessage = (message) => {
  const messageDiv = document.querySelector('#messages');
  messageDiv.innerHTML += "<br>" + message;
}

// Initialize an instance of Stripe
const stripe = Stripe("pk_test_51JI1BxKfgu58p9BHA0YrRz3WijEFpGiS9o566xzo08TAGmyBqkQXiFTS7gOtfTZa7TqSYgMwjUQXxYFMDhMc612A00KoSvxMRp");

// Create and mount the single line card element
const elements = stripe.elements();
const cardElement = elements.create('card');
cardElement.mount('#card-element');

// Extract the client secret query string argument. This is
// required to confirm the payment intent from the front-end.
const params = new URLSearchParams(window.location.search);
const clientSecret = params.get('clientSecret');

// This sample only supports a Subscription with payment
// upfront. If you offer a trial on your subscription, then
// instead of confirming the subscription's latest_invoice's
// payment_intent. You'll use stripe.confirmCardSetup to confirm
// the subscription's pending_setup_intent.
// See https://stripe.com/docs/billing/subscriptions/trials

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
  const nameInput = document.getElementById('name');

  // Create payment method and confirm payment intent.
  stripe.confirmCardPayment(clientSecret, {
    payment_method: {
      card: cardElement,
      billing_details: {
        name: nameInput.value,
      },
    }
  }).then((result) => {
    if(result.error) {
      setMessage(`Payment failed: ${result.error.message}`);
    } else {
      setMessage('Success! You can redirect to another page.');
    }
  });
});