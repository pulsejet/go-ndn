package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/zjkmxy/go-ndn/pkg/encoding/codegen"
)

var (
	inputPath   = flag.String("input", ".", "the directory containing source files")
	outputPath  = flag.String("output", "zz_generated.go", "output file path")
	packageName = flag.String("package", "", "package name of the output file")
)

// Usage is a replacement usage function for the flags package.
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of gondn_tlv_gen:\n")
	fmt.Fprintf(os.Stderr, "\tgondn_tlv_gen [-input InputPath] [-output OutputFile] [-package PackageName]\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = Usage
	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil || len(*inputPath) == 0 || len(*outputPath) == 0 {
		Usage()
		return
	}

	g := codegen.NewGenerator()
	outFullName := filepath.Join(*outputPath)
	pkgFullPath := filepath.Join(*inputPath)

	os.Remove(outFullName)

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, pkgFullPath, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	pkgName := *packageName
	for _, pkg := range pkgs {
		if strings.HasSuffix(pkg.Name, "_test") {
			continue
		} else if pkgName == "" {
			pkgName = pkg.Name
		} else if pkgName != pkg.Name {
			continue
		}
		log.Printf("processing package %s:\n", pkgName)
		for fileName, astFile := range pkg.Files {
			if filepath.Join(pkgFullPath, fileName) == outFullName {
				continue
			}
			log.Printf("\tfile %s ...\n", fileName)
			ast.Inspect(astFile, g.ProcessDecl)
		}
	}

	g.Generate(pkgName)
	result := g.Result(outFullName)
	outFile, err := os.Create(outFullName)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()
	outFile.Write(result)
}
