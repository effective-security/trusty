# Martini API

`martini` provides access to Martini API.

```.sh                                                            
usage: martini [<flags>] <command> [<args> ...]

A command-line utility for Martini API.

Flags:
      --help               Show context-sensitive help (also try --help-long and --help-man).
      --version            Show application version.
  -V, --verbose            verbose output
  -D, --debug              redirect logs to stderr
  -c, --tls-cert=TLS-CERT  client certificate for TLS connection
  -k, --tls-key=TLS-KEY    key file for client certificate
  -r, --tls-trusted-ca=""  trusted CA certificate file for TLS connection
  -s, --server=SERVER      URL of the server to control
      --retries=0          number of retries for connect failures
      --timeout=6          timeout in seconds
      --json               print responses as JSON

Commands:
  help [<command>...]
    Show help.

  status
    show the server status

  version
    show the server version

  caller
    show the caller info

  login
    login and obtain authorization token

  userinfo
    show the user profile

  orgs
    show the user's orgs

  certificates
    show the user's certificates

  subscriptions
    show the user's subscriptions

  products
    show the available products

  opencorps --name=NAME [<flags>]
    search open corporations

  fcc frn --filer=FILER
    returns FRN for filer

  fcc contact --frn=FRN
    returns contact for FRN

  org register --filer=FILER
    registers organization

  org approve --token=TOKEN --code=CODE
    approve organization validation

  org deny --token=TOKEN
    deny organization validation

  org info --token=TOKEN
    info organization request

  org validate --id=ID
    approve organization validation

  org subscribe --id=ID --product=PRODUCT
    subscribe to org

  org keys --id=ID
    list API keys

  org delete --id=ID
    delete organization

  org get --id=ID
    show the organization

  org pay --stripe-key=STRIPE-KEY --client-secret=CLIENT-SECRET [<flags>]
    pay for org

  org search [<flags>]
    search organization

  members list --id=ID
    list members

  members add --id=ID --email=EMAIL --role=ROLE
    add a member to an organization

  members remove --id=ID --member=MEMBER
    remove a member from an organization
```

For testing, the caller has to specify a trusted root certificate by providing
`-r /tmp/trusty/certs/trusty_root_ca.pem` option.
Otherwise, the test root certificate can be added to the system root store,
but this is not recommended for security reasons.

## Authentication

To start the flow, a user must authenticate.

```.sh
bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_root_ca.pem login
```

then export the token obtained from the previous command:

```.sh
export TRUSTY_AUTH_TOKEN=eyJhXXX
```

## Organizations

```.sh
bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_root_ca.pem orgs
```

## Start Organization registration flow

```.sh
bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_root_ca.pem org register --filer=123456
{
        "org": {
                "approver_email": "denis+test@martinisecurity.com",
                "approver_name": "John Doe",
                "billing_email": "denis@ekspand.com",
                "city": "PELHAM",
                "company": "TEST COMMUNICATIONS LLC",
                "created_at": "2021-08-18T13:20:00.219853Z",
                "email": "denis+test@martinisecurity.com",
                "expires_at": "2021-08-22T13:20:00.219853Z",
                "extern_id": "0123111",
                "id": "86328843001921862",
                "login": "0123111",
                "name": "TEST COMMUNICATIONS LLC",
                "phone": "2051234567",
                "postal_code": "35124",
                "provider": "martini",
                "region": "AL",
                "status": "payment_pending",
                "street_address": "241 APPLEGATE TRACE",
                "updated_at": "2021-08-18T13:20:00.219853Z"
        }
}% 
```
## Subscribe

```.sh
bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_root_ca.pem org subscribe --product prod_K3K7AZCkE3E0nF --id 82923411415760996

{
        "client_secret": "pi_3JPe0NKfgu58p9BH1Lu3Xqrr_secret_bYkh64vZLXHubuueYobMvYKnS",
        "subscription": {
                "created_at": "2021-08-18T01:55:47.827818Z",
                "currency": "usd",
                "expires_at": "2022-08-18T01:55:47.827818Z",
                "org_id": "86257988775444491",
                "price": 100,
                "status": "payment_pending"
        }
}
```

## Pay for org

```.sh
bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_root_ca.pem org pay --stripe-key <stripe publishable key> --client-secret <client secret from the response of subscribe command>

This will open a browser page where you can choose to enter payment method (card) details to make a payment.

```

## Submit for Organization validation

```.sh
bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_root_ca.pem org validate --id 82923411415760996
{
        "approver": {
                "business_name": "Low Latency Communications, LLC",
                "business_type": "Private Sector, Limited Liability Corporation",
                "contact_address": "241 Applegate Trace, Pelham, AL 35124-2945, United States",
                "contact_email": "denis+test@ekspand.com",
                "contact_fax": "",
                "contact_name": "Mr Matthew D Hardeman",
                "contact_organization": "Low Latency Communications, LLC",
                "contact_phone": "",
                "contact_position": "Secretary",
                "frn": "99999999",
                "last_updated": "",
                "registration_date": "09/29/2015 09:58:00 AM"
        },
        "code": "210507",
        "org": {
                "approver_email": "denis+test@ekspand.com",
                "approver_name": "Mr Matthew D Hardeman",
                "billing_email": "denis@ekspand.com",
                "city": "PELHAM",
                "company": "LOW LATENCY COMMUNICATIONS LLC",
                "created_at": "2021-07-26T01:30:04.813442Z",
                "email": "denis+test@ekspand.com",
                "expires_at": "2021-07-30T01:30:04.813442Z",
                "extern_id": "99999999",
                "id": "82923411415760996",
                "login": "99999999",
                "name": "LOW LATENCY COMMUNICATIONS LLC",
                "phone": "2057453970",
                "postal_code": "35124",
                "provider": "martini",
                "region": "AL",
                "status": "validation_pending",
                "street_address": "241 APPLEGATE TRACE",
                "updated_at": "2021-07-26T01:30:04.813442Z"
        }
}
```

## Approve Organization registration

```.sh
bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_root_ca.pem org approve --token nNwZipSV2rAPkbsZ --code 191161
{
        "org": {
                "approver_email": "denis+test@ekspand.com",
                "approver_name": "Mr Matthew D Hardeman",
                "billing_email": "denis@ekspand.com",
                "city": "PELHAM",
                "company": "LOW LATENCY COMMUNICATIONS LLC",
                "created_at": "2021-07-26T01:30:04.813442Z",
                "email": "denis+test@ekspand.com",
                "expires_at": "2021-07-30T01:30:04.813442Z",
                "extern_id": "99999999",
                "id": "82923411415760996",
                "login": "99999999",
                "name": "LOW LATENCY COMMUNICATIONS LLC",
                "phone": "2057453970",
                "postal_code": "35124",
                "provider": "martini",
                "region": "AL",
                "status": "approved",
                "street_address": "241 APPLEGATE TRACE",
                "updated_at": "2021-07-26T01:30:04.813442Z"
        }
}
```

## Get API Keys

```.sh
bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_root_ca.pem org keys --id 82936541768319076                  
{
        "keys": [
                {
                        "billing": false,
                        "created_at": "2021-07-26T03:41:34.618784Z",
                        "enrollment": true,
                        "expires_at": "2021-07-30T03:40:31.112741Z",
                        "id": "82936648303640676",
                        "key": "_0zxP8c4AUrj_vnPmGXU_eEbA3AzkTXZ",
                        "management": false,
                        "org_id": "82936541768319076",
                        "used_at": null
                }
        ]
}
```

## Get Org members

```.sh
bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_root_ca.pem org members --id 82936541768319076                  
{
    "members": [
        {
            "email": "denis@ekspand.com",
            "membership_id": "85334042257457478",
            "name": "Denis Issoupov",
            "org_id": "85334042257391942",
            "org_name": "LOW LATENCY COMMUNICATIONS LLC",
            "role": "admin",
            "source": "martini",
            "user_id": "85232539848933702"
        }
    ]
}
```

## List Certificates

```.sh
bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_root_ca.pem certificates
```

## List Subscriptions

```.sh
bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_root_ca.pem subscriptions
```

## List Products

```.sh
bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_root_ca.pem products
```

## Search org

```.sh
 bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_root_ca.pem org search --frn 0123111 --filler 123111
{
        "orgs": [
                {
                        "approver_email": "denis@martinisecurity.com",
                        "approver_name": "John Doe",
                        "billing_email": "denis@martinisecurity.com",
                        "city": "PELHAM",
                        "company": "TEST COMMUNICATIONS LLC",
                        "created_at": "2021-09-02T00:15:22.725605Z",
                        "email": "denis@martinisecurity.com",
                        "expires_at": "2024-09-02T00:16:02.915704Z",
                        "extern_id": "0123111",
                        "id": "88424187273675065",
                        "login": "0123111",
                        "name": "TEST COMMUNICATIONS LLC",
                        "phone": "2051234567",
                        "postal_code": "35124",
                        "provider": "martini",
                        "region": "AL",
                        "registration_id": "123111",
                        "status": "approved",
                        "street_address": "241 APPLEGATE TRACE",
                        "updated_at": "2021-09-02T00:15:22.725605Z"
                }
        ]
}
```