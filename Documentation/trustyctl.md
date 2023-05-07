# trustyctl

```.sh
usage: trustyctl [<flags>] <command> [<args> ...]

A command-line utility for controlling Trusty.

Flags:
      --help               Show context-sensitive help (also try --help-long and --help-man).
      --version            Show application version.
  -V, --verbose            verbose output
  -D, --debug              redirect logs to stderr
      --cfg="trusty-config.yaml"
                           trusty configuration file
      --hsm-cfg=HSM-CFG    HSM provider configuration file
      --crypto-prov=CRYPTO-PROV ...
                           path to additional Crypto provider configurations
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

  ca issuers [<flags>]
    show the issuing CAs

  ca profile [<flags>]
    show the certificate profile

  ca sign --csr=CSR --profile=PROFILE [<flags>]
    sign CSR

  ca certs --ikid=IKID [<flags>]
    print the certificates

  ca label [<flags>]
    update the certificate label

  ca revoked --ikid=IKID [<flags>]
    print the revoked certificates

  ca publish_crl --ikid=IKID
    publish CRL

  ca revoke [<flags>]
    revoke a certificate

  ca certificate [<flags>]
    get a certificate

  cis roots [<flags>]
    show the roots
```

## Export the host to be used

  export TRUSTY_SERVER=https://localhost:7892


## Server status

```.sh
bin/trustyctl status

  Name        | ca                         
  Node        |                            
  Host        | dissoupov-wsl2             
  Listen URLs | https://0.0.0.0:7892       
  Version     | 0.1.57-dissoupov-ltl2      
  Runtime     | go1.17                     
  Started     | 2021-11-25T10:23:34-08:00  
  Uptime      | 10s 
```

## Server version

```.sh
bin/trustyctl -s https://localhost:7892 version

0.1.57-dissoupov-ltl2 (go1.17) 
```

## Caller identity

```.sh
bin/trustyctl -s https://localhost:7892 caller

  Name | trusty             
  ID   |                    
  Role | authenticated_tls 
```

## Issuers list

```.sh
bin/trustyctl -s https://localhost:7892 ca issuers

=========================================================
Label: trusty.svc
Profiles: [codesign default client ocsp test_server timestamp server peer test_client]
Subject: C=US, L=WA, O=trusty.com, CN=[TEST] Trusty Level 2 CA
  Issuer: C=US, L=WA, O=trusty.com, CN=[TEST] Trusty Level 1 CA
  SKID: e91a37cee2b390936e2b850193819c95028e8372
  IKID: bc6701e1bac3b9b4b1033dbe360272d0365febf5
  Serial: 667451990703804007427228107864054685450398864260
  Issued: 2021-11-25 09:28:00 -0800 PST (1h0m0s ago)
  Expires: 2026-11-24 09:28:00 -0800 PST (in 43798h59m0s)
```

