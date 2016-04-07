static-server
=============

## About

The `static-server` utility is useful for testing. It serves the contents of a
given directory over http.

    $ static-server -h
    Usage:
      static-server-netgo [OPTIONS]
    
    Application Options:
      -r, --root=       Root directory to server from (default: .)
      -l, --listen=     Address:Port to bind to for HTTP (default: 0.0.0.0:8000)
          --ssl-listen= Address:Port to bind to for HTTPS/SSL/TLS
          --ssl-key=    ssl private key (key.pem) path
          --ssl-cert=   ssl cert (cert.pem) path
      -v, --verbose     Show verbose (debug) log level output
      -V, --version     Print version and exit; specify twice to show license information
    
    Help Options:
      -h, --help        Show this help message
    

## License

Released under the [MIT
license](http://www.opensource.org/licenses/mit-license.php). See `LICENSE.md`
file for details.
