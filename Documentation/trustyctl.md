# trustyctl

```.sh
Usage: trustyctl <command>

CTL for trusty server

Flags:
  -h, --help                                  Show context-sensitive help.
  -D, --debug                                 Enable debug mode
  -l, --log-level="critical"                  Set the logging level (debug|info|warn|error|critical)
      --o=STRING                              Print output format
      --cfg="~/.config/trusty/config.yaml"    Service configuration file
  -s, --server=STRING                         Address of the remote server to connect. Use TRUSTY_SERVER environment to override
  -c, --cert=STRING                           Client certificate file for mTLS
  -k, --cert-key=STRING                       Client certificate key for mTLS
  -r, --trusted-ca=STRING                     Trusted CA store for server TLS
  -t, --timeout=6                             Timeout in seconds

Commands:
  version               print remote server version
  status                print remote server status
  caller                print identity of the current user
  ca issuers            list issuers certificates
  ca certs              list certificates
  ca revoked            list revoked certificates
  ca profile            show certificate profile
  ca sign               sign certificate
  ca publish-crl        publish CRL
  ca revoke             revoke certificate
  ca set-cert-label     set certificate label
  ca get-certificate    get certificate
  cis roots             list Root certificates

Run "trustyctl <command> --help" for more information on a command.
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

