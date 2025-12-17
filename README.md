static-server
=============

> [!IMPORTANT]
> Moved to [codeberg](https://codeberg.org/dropwhile/static-server)!

## About

The `static-server` utility is useful for testing. It serves the contents of a
given directory over http.


## Install

```
go install codeberg.org/dropwhile/static-server/cmd/static-server@latest
```

## Usage

```
Usage:
  static-server [OPTIONS]

Application Options:
      --indexes=     comma separated (ordered) list of index files (default: index.html)
      --readmes=     comma separated (ordered) list of readme files to auto append to dir listings
      --headers=     comma separated (ordered) list of header files to auto prepend to dir listings
  -t, --template=    template file to use for directory listings
  -r, --root=        Root directory to server from (default: .)
      --no-log-ts    Do not add a timestamp to logging
      --log-json     Log messages in json format
      --log-struct   Log messages in structured text format
  -l, --listen=      Address:Port to bind to for HTTP (default: 0.0.0.0:8000)
      --ssl-listen=  Address:Port to bind to for HTTPS/SSL/TLS
      --ssl-key=     ssl private key (key.pem) path
      --ssl-cert=    ssl cert (cert.pem) path
  -x, --no-indexing  disable directory indexing
  -v, --verbose      Show verbose (debug) log level output
  -V, --version      Print version and exit; specify twice to show license information

Help Options:
  -h, --help         Show this help message
```

## License

Released under the [ISC license][1] . See `LICENSE.md` file for details.

[1]: https://choosealicense.com/licenses/isc/
