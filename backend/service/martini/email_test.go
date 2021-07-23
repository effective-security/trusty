package martini

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const userEmailTemplate = `
<h1>
    You will need to provide the contact this number fo complete the validation.
    <h2>{{.Number}}</h2>
</h1>
`

const contactEmailTemplate = `
<h1>
    {{.Email}} has requested permission to acquire certificates for your organization.
    <h2>{{.OrgDetails}}</h2>
    If this request was authorized, enter the number that {{.Email}} provided you below and click "Approve".
</h1>
`

func Test_renderEmailTemplate(t *testing.T) {

	userEmailData := struct {
		Number string
	}{
		Number: "123456",
	}

	body, err := renderEmailTemplate(userEmailTemplate, userEmailData)
	require.NoError(t, err)
	require.Equal(t, "\n<h1>\n    You will need to provide the contact this number fo complete the validation.\n    <h2>123456</h2>\n</h1>\n", body)
	contactEmailData := struct {
		Email      string
		OrgDetails string
	}{
		Email:      "hayk.baluyan@gmail.com",
		OrgDetails: "My Organization",
	}

	body, err = renderEmailTemplate(contactEmailTemplate, contactEmailData)
	require.NoError(t, err)
	require.Equal(t, "\n<h1>\n    hayk.baluyan@gmail.com has requested permission to acquire certificates for your organization.\n    <h2>My Organization</h2>\n    If this request was authorized, enter the number that hayk.baluyan@gmail.com provided you below and click \"Approve\".\n</h1>\n", body)
}
