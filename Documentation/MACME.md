# Martini API

`macme` provides ACME client.

```.sh                                                            
usage: macme [<flags>] <command> [<args> ...]

Martini ACME client

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

  acme directory
    show ACME directory

  acme account --id=ID --key=KEY
    show registered account

  acme register --id=ID --key=KEY --contact=CONTACT
    register account

  acme order --id=ID --spc=SPC [<flags>]
    order certificate
```

For testing, the caller has to specify a trusted root certificate by providing
`-r /tmp/trusty/certs/trusty_root_ca.pem` option.
Otherwise, the test root certificate can be added to the system root store,
but this is not recommended for security reasons.

## Register ACME account

```.sh
bin/macme -s https://localhost:7891 -r /tmp/trusty/certs/trusty_root_ca.pem acme register --id 88368917956788537 --key pIIwjGjpHqsx0Bgrq4wYc1eI9VIOQu80  --contact denis@ekspand.com
{
        "account_url": "https://localhost:7891/v2/acme/account/88376537211994425",
        "fingerprint": "SHA256 DD:2A:51:D8:35:22:77:36:02:FF:83:E1:09:00:05:D3:D3:12:25:55:4D:FB:BC:82:8C:E7:79:B0:A3:C8:20:93",
        "key_id": "88368917956788537",
        "registration": {
                "contact": [
                        "denis@ekspand.com"
                ],
                "orders": "https://localhost:7891/v2/acme/account/88376537211994425/orders",
                "status": "valid",
                "termsOfServiceAgreed": true
        }
}
```

## Request certificate 

```.sh
 bin/macme -s https://localhost:7891 -r /tmp/trusty/certs/trusty_root_ca.pem acme order --id 84350577525391460 --spc /tmp/spc                         

certificate: /home/dissoupov/.mrtsec/certificates/ddadd515eb4e758f7a8a18a4093574dacdac4cf2.pem
key: /home/dissoupov/.mrtsec/certificates/ddadd515eb4e758f7a8a18a4093574dacdac4cf2.key

```
