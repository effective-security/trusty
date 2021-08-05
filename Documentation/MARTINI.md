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

  login
    login and obtain authorization token

  userinfo
    show the user profile

  orgs
    show the user's orgs

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

  org validate --org=ORG
    approve organization validation

  org subscribe --org=ORG --cardholder=CARDHOLDER --cc=CC --expiry=EXPIRY --cvv=CVV --years=YEARS
    create subscription

```

For testing, the caller has to specify a trusted root certificate by providing
`-r /tmp/trusty/certs/trusty_dev_root_ca.pem` option.
Otherwise, the test root certificate can be added to the system root store,
but this is not recommended for security reasons.

## Authentication

To start the flow, a user must authenticate.

```.sh
bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_dev_root_ca.pem login
```

then export the token obtained from the previous command:

```.sh
export TRUSTY_AUTH_TOKEN=eyJhXXX
```

## Organizations

```.sh
bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_dev_root_ca.pem orgs
```

## Start Organization registration flow

```.sh
bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_dev_root_ca.pem org register --filer=123456
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
        "code": "191161",
        "org": {
                "approver_email": "denis+test@ekspand.com",
                "approver_name": "Mr Matthew D Hardeman",
                "billing_email": "denis@ekspand.com",
                "city": "PELHAM",
                "company": "LOW LATENCY COMMUNICATIONS LLC",
                "created_at": "2021-07-25T13:52:57.081861Z",
                "email": "denis+test@ekspand.com",
                "expires_at": "2022-07-23T01:52:57.081861Z",
                "extern_id": "99999999",
                "id": "82853236129661028",
                "login": "99999999",
                "name": "LOW LATENCY COMMUNICATIONS LLC",
                "phone": "2057453970",
                "postal_code": "35124",
                "provider": "martini",
                "region": "AL",
                "status": "pending",
                "street_address": "241 APPLEGATE TRACE",
                "updated_at": "2021-07-25T13:52:57.081861Z"
        }
}
```
## Subscribe

```.sh
bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_dev_root_ca.pem org subscribe --org 82923411415760996 --cardholder "Denis Issoupov" --cc 4445-1234-1234-1234 --expiry 11/22 --cvv 266 --years 3

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
                "status": "validation_pending",
                "street_address": "241 APPLEGATE TRACE",
                "updated_at": "2021-07-26T01:30:04.813442Z"
        }
}
```

## Submit for Organization validation

```.sh
bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_dev_root_ca.pem org validate --org 82923411415760996
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
bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_dev_root_ca.pem org approve --token nNwZipSV2rAPkbsZ --code 191161
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
bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_dev_root_ca.pem org keys --org 82936541768319076                  
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

## Register ACME account

```.sh
bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_dev_root_ca.pem acme register --org 84350577525391460 --key sJgHnIOI96okZ9+7N4OfqZl/27V/phF/ --contact denis@ekspand.com
{
        "account_url": "https://localhost:7891/v2/acme/account/84350771084132452",
        "fingerprint": "SHA256 CC:9D:59:F2:92:91:76:FF:37:C1:60:FC:91:FC:BB:B2:A7:7C:7F:93:D7:D4:26:9D:AA:B9:82:BC:8D:EE:11:25",
        "org_id": "84350577525391460",
        "registration": {
                "contact": [
                        "denis@ekspand.com"
                ],
                "orders": "https://localhost:7891/v2/acme/account/84350771084132452/orders",
                "status": "valid",
                "termsOfServiceAgreed": true
        }
}
```

## Request certificate 

```.sh
 bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_dev_root_ca.pem acme order --org 84350577525391460 --spc /tmp/spc                                                       

certificate: /home/dissoupov/.mrtsec/certificates/ddadd515eb4e758f7a8a18a4093574dacdac4cf2.pem
key: /home/dissoupov/.mrtsec/certificates/ddadd515eb4e758f7a8a18a4093574dacdac4cf2.key

```

## List Certificates

```.sh
bin/martini -s https://localhost:7891 -r /tmp/trusty/certs/trusty_dev_root_ca.pem certificates
```
