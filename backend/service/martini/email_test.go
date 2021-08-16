package martini

import (
	"testing"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderUserEmailTemplate(t *testing.T) {
	emailData := orgValidationEmailTemplate{
		RequesterName:  "Hayk Baluyan",
		RequesterEmail: "hayk.baluyan@gmail.com",
		ApproverName:   "Denis Issoupov",
		ApproverEmail:  "denis@ekspand.com",
		Hostname:       "app.dev.martinisecurity.com",
		Code:           "123456",
		Token:          "abcd-1234-5678",
		Company:        "IPIPHONY",
		Address:        "123 Drive, Kirkland, 98034, WA",
	}

	body, err := renderEmailTemplate(requesterEmailTemplate2, emailData)
	require.NoError(t, err)
	assert.Equal(t, `
<h2>Organization validation submitted</h2>
<p>
	<div>Hayk Baluyan,</div>
	<div>The organization validation request has been sent to Denis Issoupov, denis@ekspand.com.</div>

    <div>Please provide the approver this code fo complete the validation.</div>
    <h3>123456</h3>

	<div>Thank you for using Martini Security!</div>
</p>
`, body)

	body, err = renderEmailTemplate(approverEmailTemplate2, emailData)
	require.NoError(t, err)
	assert.Equal(t, `
<h2>Organization validation request</h2>
<p>
	<div>Denis Issoupov,</div>

    <div>Hayk Baluyan, hayk.baluyan@gmail.com has requested permission to acquire certificates for your organization.</div>

    <h2>IPIPHONY</h2>
	<h4>123 Drive, Kirkland, 98034, WA</h4>
	
    <div>To authorize this request, enter the Code that was provided you by the requester.</div>
	<h3>Link: <a href="https://app.dev.martinisecurity.com/validate/abcd-1234-5678">Click here to approve</a></h3>

	<div>Thank you for using Martini Security!</div>
</p>
`, body)
}

func TestRenderApprovedTemplate(t *testing.T) {
	emailData := v1.Organization{
		Company: "IPIPHONY",
	}

	body, err := renderEmailTemplate(orgApprovedTemplate2, emailData)
	require.NoError(t, err)
	assert.Equal(t, `
<h2>Organization validation succeeded!</h2>
<p>
	<div>IPIPHONY is approved to request certificates.</div>

	<div>Thank you for using Martini Security!</div>
</p>
`, body)
}

const requesterEmailTemplate2 = `
<h2>Organization validation submitted</h2>
<p>
	<div>{{.RequesterName}},</div>
	<div>The organization validation request has been sent to {{.ApproverName}}, {{.ApproverEmail}}.</div>

    <div>Please provide the approver this code fo complete the validation.</div>
    <h3>{{.Code}}</h3>

	<div>Thank you for using Martini Security!</div>
</p>
`

const approverEmailTemplate2 = `
<h2>Organization validation request</h2>
<p>
	<div>{{.ApproverName}},</div>

    <div>{{.RequesterName}}, {{.RequesterEmail}} has requested permission to acquire certificates for your organization.</div>

    <h2>{{.Company}}</h2>
	<h4>{{.Address}}</h4>
	
    <div>To authorize this request, enter the Code that was provided you by the requester.</div>
	<h3>Link: <a href="https://{{.Hostname}}/validate/{{.Token}}">Click here to approve</a></h3>

	<div>Thank you for using Martini Security!</div>
</p>
`

const orgApprovedTemplate2 = `
<h2>Organization validation succeeded!</h2>
<p>
	<div>{{.Company}} is approved to request certificates.</div>

	<div>Thank you for using Martini Security!</div>
</p>
`
