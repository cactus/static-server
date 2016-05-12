// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"html/template"
	"strings"
	"time"
)

// DirItem represents a directory item for templating
type DirItem struct {
	Name    string
	Link    string
	Size    string
	ModTime time.Time
}

// JustPadding returns a string with size elements of pre-padding
func JustPadding(s string, size int) string {
	return strings.Repeat(" ", size-len(s))
}

// Page represents a page for templating
type Page struct {
	Path       string
	DirItems   []*DirItem
	ReadmeText template.HTML
	HeaderText template.HTML
}

var tplFuncMap = template.FuncMap{
	"justpadding": JustPadding,
}

const tplDefault = `
<html>
{{with .Path}}<head><title>Index of /{{.}}</title></head>{{end}}
<body bgcolor="white">
{{with .Path}}<h1>Index of /{{.}}</h1>{{end}}
{{- with .HeaderText -}}<p>{{.}}</p>{{end}}
<hr>
<pre><a href="../">../</a>
{{range $item := .DirItems}}<a href="{{$item.Link}}">{{$item.Name}}</a>{{justpadding $item.Name 50}} {{.ModTime.Format "2006-01-02 15:04"}} {{.Size | printf "%20s"}}
{{end}}</pre>
<hr>
{{- with .ReadmeText -}}<p>{{.}}</p>{{end}}
</body>
</html>
`
