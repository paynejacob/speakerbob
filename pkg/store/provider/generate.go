package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
)

type Resource struct {
	Name          string
	Package       string
	KeyField      string
	IndexedFields []string
}

//go:embed provider.go.tmpl
var templateFS embed.FS

var indexedFieldsRaw string

func main() {
	var err error
	var dest *os.File
	var resource Resource

	t := template.Must(template.ParseFS(templateFS, "*.tmpl"))

	flag.StringVar(&resource.Name, "type", "", "the provider type")
	flag.StringVar(&resource.Package, "package", "", "the package of the type")
	flag.StringVar(&resource.KeyField, "key_field", "Id", "the field name to be used as the key")
	flag.StringVar(&indexedFieldsRaw, "indexed_fields", "", "comma separated list of fields to index")
	flag.Parse()

	resource.IndexedFields = strings.Split(indexedFieldsRaw, ",")

	filename := fmt.Sprintf("zz_%s_provider.go", resource.Name)

	_ = os.Remove(filename)
	dest, err = os.Create(filename)
	if err != nil {
		log.Fatal(err.Error())
	}

	defer dest.Close()

	err = t.Execute(dest, resource)
	if err != nil {
		log.Fatal(err.Error())
	}
}
