// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Some parts from go-std-lib/src/net/http/fs.go
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found at https://golang.org/LICENSE

package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

var unixEpochTime = time.Unix(0, 0)

type fileServer struct {
	root       http.FileSystem
	mainTpl    *template.Template
	indexList  []string
	headerList []string
	readmeList []string
}

func (fs *fileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}
	fs.serveFile(w, r, path.Clean(upath))
}

func (fs *fileServer) dirList(w http.ResponseWriter, f http.File) {
	dirs, err := f.Readdir(-1)
	if err != nil {
		http.Error(w, "Error reading directory", http.StatusInternalServerError)
		return
	}
	sort.Sort(byName(dirs))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fname := ""
	fst, err := f.Stat()
	if err == nil {
		fname = fst.Name() + "/"
	}

	page := &Page{Path: fname, DirItems: make([]*DirItem, 0)}

	dirItems := make([]*DirItem, 0)
	fileItems := make([]*DirItem, 0)

	for _, d := range dirs {
		name := d.Name()
		size := ""
		isdir := false
		if d.IsDir() {
			isdir = true
			name += "/"
			size = "-"
		} else {
			size = strconv.FormatInt(d.Size(), 10)
		}

		// name may contain '?' or '#', which must be escaped to remain
		// part of the URL path, and not indicate the start of a query
		// string or fragment.
		url := url.URL{Path: name}

		if len(name) > 50 {
			name = name[:47] + "..>"
		}

		if isdir {
			dirItems = append(dirItems, &DirItem{
				Link:    url.String(),
				Name:    name,
				Size:    size,
				ModTime: d.ModTime(),
			})
		} else {
			fileItems = append(fileItems, &DirItem{
				Link:    url.String(),
				Name:    name,
				Size:    size,
				ModTime: d.ModTime(),
			})
		}
	}

	page.DirItems = append(page.DirItems, dirItems...)
	page.DirItems = append(page.DirItems, fileItems...)

	if fname != "" {
		// find header/footer if present
		for _, h := range fs.headerList {
			dn := path.Join(fname, h)
			data, err := ioutil.ReadFile(dn)
			if err == nil {
				page.HeaderText = template.HTML(data)
			}
		}

		for _, r := range fs.readmeList {
			dn := path.Join(fname, r)
			data, err := ioutil.ReadFile(dn)
			if err == nil {
				page.ReadmeText = template.HTML(data)
			}
		}
	}

	fs.mainTpl.Execute(w, page)
}

// name is '/'-separated, not filepath.Separator.
func (fs *fileServer) serveFile(w http.ResponseWriter, r *http.Request, name string) {
	f, d, err := openWithStat(fs.root, name)
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}
	defer f.Close()

	// redirect to canonical path: / at end of directory url
	// r.URL.Path always begins with /
	url := r.URL.Path
	if d.IsDir() {
		if url[len(url)-1] != '/' {
			localRedirect(w, r, path.Base(url)+"/")
			return
		}
	} else {
		// path.Clean always strips the trailing slash, so even though
		// we may have a file (eg. /index.html/), the url may in fact
		// be invalid anyway.
		if url[len(url)-1] == '/' {
			http.Error(w, "404 page not found", http.StatusNotFound)
			return
		}

		// no symlinks or special files
		if !d.Mode().IsRegular() {
			http.Error(w, "403 Forbidden", http.StatusForbidden)
		}
	}

	// use contents of index page for directory, if present
	if d.IsDir() {
		for _, indexPage := range fs.indexList {
			index := strings.TrimSuffix(name, "/") + "/" + indexPage
			ff, dd, err := openWithStat(fs.root, index)
			if err == nil {
				defer ff.Close()
				http.ServeContent(w, r, dd.Name(), dd.ModTime(), ff)
				return
			}
		}

		// didn't find an index file
		if checkLastModified(w, r, d.ModTime()) {
			return
		}
		fs.dirList(w, f)
		return
	}

	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
}

func openWithStat(fs http.FileSystem, name string) (http.File, os.FileInfo, error) {
	f, err := fs.Open(name)
	if err != nil {
		return nil, nil, err
	}

	d, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, nil, err
	}

	return f, d, nil
}

func staticServer(root http.FileSystem, tpl *template.Template, indexList, headerList, readmeList []string) http.Handler {
	if len(indexList) == 0 {
		indexList = []string{"index.html"}
	}

	if tpl == nil {
		tpl = template.Must(template.New("main").Funcs(tplFuncMap).Parse(strings.TrimSpace(tplDefault)))
	}

	return &fileServer{
		root:       root,
		mainTpl:    tpl,
		indexList:  indexList,
		readmeList: readmeList,
		headerList: headerList,
	}
}
