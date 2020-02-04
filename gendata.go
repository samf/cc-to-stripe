// +build ignore

package main

import (
	"log"

	"github.com/samf/cc-to-stripe/assets"

	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(assets.Assets, vfsgen.Options{
		Filename:     "assets/assets.go",
		PackageName:  "assets",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
