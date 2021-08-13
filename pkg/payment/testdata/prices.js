const createSubscription = (orgID,productID) => {
  const trustyAuthToken = document.getElementById('trustyAuthToken').value;
  return fetch('https://localhost:7891/v1/ms/subscription/create', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer '+ trustyAuthToken,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      org_id: orgID,
      product_id: "prod_K2OpdTIt5JQxoW",
    }),
  })
    .then((response) => response.json())
    .then((data) => {
      const params = new URLSearchParams(window.location.search);
      params.append('clientSecret', data.client_secret);
      window.location.href = 'subscribe.html?' + params.toString();
    })
    .catch((error) => {
      console.error('Error:', error);
    });
}