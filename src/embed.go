package main

import (
	_ "embed"
)

//go:embed templates/gitignore
var gitignoreTemplate string

//go:embed embed/apache-2.0.txt
var apache20Fallback string
