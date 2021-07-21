# Martini API

`martinictl` provides access to Martini API.

```.sh                                                            
usage: martinictl [<flags>] <command> [<args> ...]

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
```

## Authentication

To start the flow, a user must authenticate.

```.sh
bin/martinictl -s https://localhost:7891 -r /tmp/trusty/certs/trusty_dev_root_ca.pem login
```

then export the token obtained from the previous command:

```.sh
export TRUSTY_AUTH_TOKEN=eyJhXXX
```

## Organizations

```.sh
bin/martinictl -s https://localhost:7891 -r /tmp/trusty/certs/trusty_dev_root_ca.pem orgs
```