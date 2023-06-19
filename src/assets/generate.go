//go:build ignore
// +build ignore

// Copyright indece UG (haftungsbeschränkt) - All rights reserved.
//
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential
//
// Written by Stephan Lukas <stephan.lukas@indece.com>, 2022

package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/shurcooL/vfsgen"
)

func main() {
	var cwd, _ = os.Getwd()
	assets := http.Dir(filepath.Join(cwd, "../assets"))

	err := vfsgen.Generate(assets, vfsgen.Options{
		Filename:     "assets/assets_vfsdata.gen.go",
		PackageName:  "assets",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
