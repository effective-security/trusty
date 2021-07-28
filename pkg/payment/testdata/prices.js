const createSubscription = (orgID,years) => {
  const trustyAuthToken = document.getElementById('trustyAuthToken').value;
  return fetch('https://localhost:7891/v1/ms/subscription/create', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer '+ trustyAuthToken,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      org_id: orgID,
      years: years,
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