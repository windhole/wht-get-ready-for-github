package main

import (
	_ "embed"
)

//go:embed embed/gitignore.tmpl
var gitignoreTemplate string

//go:embed embed/apache-2.0.txt
var apache20Fallback string
