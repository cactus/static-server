// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// static-server daemon
package main

//go:generate go run ../../tools/genversion.go -pkg $GOPACKAGE -input ../../DEPS.md -output version_info_generated.go

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	flags "github.com/jessevdk/go-flags"
)

var (
	ServerName    = "static-server"
	ServerVersion = "no-version"
)

func main() {
	// command line flags
	var opts struct {
		IndexList      string `long:"indexes" default:"index.html" description:"comma separated (ordered) list of index files"`
		ReadmeList     string `long:"readmes" default:"" description:"comma separated (ordered) list of readme files to auto append to dir listings"`
		HeaderList     string `long:"headers" default:"" description:"comma separated (ordered) list of header files to auto prepend to dir listings"`
		Template       string `long:"template" short:"t" default:"" description:"template file to use for directory listings"`
		RootDir        string `long:"root" short:"r" default:"." description:"Root directory to server from"`
		BindAddress    string `long:"listen" short:"l" default:"0.0.0.0:8000" description:"Address:Port to bind to for HTTP"`
		BindAddressSSL string `long:"ssl-listen" description:"Address:Port to bind to for HTTPS/SSL/TLS"`
		SSLKey         string `long:"ssl-key" description:"ssl private key (key.pem) path"`
		SSLCert        string `long:"ssl-cert" description:"ssl cert (cert.pem) path"`
		NoIndex        bool   `long:"no-indexing" short:"x"  description:"disable directory indexing"`
		Verbose        bool   `short:"v" long:"verbose" description:"Show verbose (debug) log level output"`
		Version        []bool `short:"V" long:"version" description:"Print version and exit; specify twice to show license information"`
	}

	// parse said flags
	_, err := flags.Parse(&opts)
	if err != nil {
		if e, ok := err.(*flags.Error); ok {
			if e.Type == flags.ErrHelp {
				os.Exit(0)
			}
		}
		os.Exit(1)
	}

	if len(opts.Version) > 0 {
		fmt.Printf("%s %s (%s,%s-%s)\n", ServerName, ServerVersion, runtime.Version(), runtime.Compiler, runtime.GOARCH)
		if len(opts.Version) > 1 {
			fmt.Printf("\n%s\n", strings.TrimSpace(licenseText))
		}
		os.Exit(0)
	}

	if opts.BindAddress == "" && opts.BindAddressSSL == "" {
		log.Fatal("One of listen or ssl-listen required")
	}

	if opts.BindAddressSSL != "" && opts.SSLKey == "" {
		log.Fatal("ssl-key is required when specifying ssl-listen")
	}
	if opts.BindAddressSSL != "" && opts.SSLCert == "" {
		log.Fatal("ssl-cert is required when specifying ssl-listen")
	}

	if finfo, err := os.Stat(opts.RootDir); os.IsNotExist(err) || finfo.Mode().IsDir() != true {
		log.Fatal("Specified root directory is not readable, not present, or not a directory")
	}

	indexes := []string{}
	for _, s := range strings.Split(opts.IndexList, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			indexes = append(indexes, s)
		}
	}

	headers := []string{}
	for _, s := range strings.Split(opts.HeaderList, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			headers = append(headers, s)
		}
	}

	readmes := []string{}
	for _, s := range strings.Split(opts.ReadmeList, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			readmes = append(readmes, s)
		}
	}

	var tpl *template.Template
	if opts.Template != "" {
		tplText, err := ioutil.ReadFile(opts.Template)
		if err != nil {
			log.Fatal(err)
		}
		tpl = template.Must(template.New("main").Funcs(tplFuncMap).Parse(strings.TrimSpace(string(tplText))))
	}

	indexing := true
	if opts.NoIndex == true {
		indexing = false
	}

	staticServer := staticServer(http.Dir(opts.RootDir), tpl, indexes, headers, readmes, indexing)

	if opts.BindAddress != "" {
		log.Println("Starting server on", opts.BindAddress)
		go func() {
			srv := &http.Server{
				Addr:        opts.BindAddress,
				ReadTimeout: 30 * time.Second,
				Handler:     staticServer,
			}
			log.Fatal(srv.ListenAndServe())
		}()
	}
	if opts.BindAddressSSL != "" {
		log.Println("Starting TLS server on", opts.BindAddressSSL)
		go func() {
			srv := &http.Server{
				Addr:        opts.BindAddressSSL,
				ReadTimeout: 30 * time.Second,
				Handler:     staticServer,
			}
			log.Fatal(srv.ListenAndServeTLS(opts.SSLCert, opts.SSLKey))
		}()
	}

	// just block. listen and serve will exit the program if they fail/return
	// so we just need to block to prevent main from exiting.
	select {}
}
