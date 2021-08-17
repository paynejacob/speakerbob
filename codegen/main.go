package main

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"github.com/fatih/structtag"
	"go/types"
	"golang.org/x/tools/go/packages"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

type resource struct {
	Name             string
	Package          string
	PackageName      string
	KeyField         string
	SearchableFields []string
	LookupFields     []struct {
		Name    string
		IsArray bool
		Type    string
	}
}

type FieldAttributes struct {
	IsKey      bool
	Searchable bool
	Lookup     bool
}

//go:embed templates/root.tmpl
var templateFS embed.FS

const tagKey = "store"

var tmpls = template.Must(template.ParseFS(templateFS, "templates/*.tmpl"))

func main() {
	// parse arguments
	args := os.Args
	if len(args) < 2 {
		log.Fatal("you must provide a struct to generate a provider for")
		return
	}

	for _, typePath := range args[1:] {
		_resource := resource{}

		_resource.Package, _resource.Name = parseSourceType(typePath)

		pkg, err := loadPackage(_resource.Package)
		if err != nil {
			log.Fatalf(err.Error())
			return
		}

		pkgParts := strings.Split(pkg.String(), "/")
		_resource.PackageName = pkgParts[len(pkgParts)-1]

		obj := pkg.Types.Scope().Lookup(_resource.Name)

		if _, ok := obj.(*types.TypeName); !ok {
			log.Fatalf("%s is not a named type", _resource.Name)
			return
		}

		t, ok := obj.Type().Underlying().(*types.Struct)
		if !ok {
			log.Fatalf("%s is not a struct", _resource.Name)
			return
		}

		// Map fields
		for i := 0; i < t.NumFields(); i++ {
			field := t.Field(i)
			fieldAttrs := parseTag(t.Tag(i))

			if fieldAttrs.IsKey {
				if _resource.KeyField != "" {
					log.Fatalf("%s has more than 1 key field", _resource.Name)
					return
				}

				_resource.KeyField = field.Name()
			}

			if fieldAttrs.Searchable {
				if field.Type().String() != "string" && field.Type().String() != "[]byte" {
					log.Fatalf("%s: searchable fields must be type [string] or [[]byte]", field.Name())
					return
				}
				_resource.SearchableFields = append(_resource.SearchableFields, field.Name())
			}

			if fieldAttrs.Lookup {
				renderType := strings.TrimPrefix(field.Type().String(), "[]")
				renderType = strings.TrimPrefix(renderType, _resource.Package)
				renderType = strings.TrimPrefix(renderType, ".")

				_resource.LookupFields = append(_resource.LookupFields, struct {
					Name    string
					IsArray bool
					Type    string
				}{Name: field.Name(), Type: renderType, IsArray: strings.HasPrefix(field.Type().String(), "[]")})
			}
		}

		if _resource.KeyField == "" {
			log.Fatalf("%s does not define a key field", _resource.Name)
			return
		}

		err = render(_resource)
		if err != nil {
			log.Fatal(err.Error())
			return
		}
	}
}

func parseTag(s string) FieldAttributes {
	fa := FieldAttributes{}

	tgs, err := structtag.Parse(s)
	if err != nil {
		return fa
	}

	tag, err := tgs.Get(tagKey)
	if err != nil {
		return fa
	}

	for _, attr := range append(tag.Options, tag.Name) {
		switch attr {
		case "lookup":
			fa.Lookup = true
		case "searchable":
			fa.Searchable = true
		case "key":
			fa.IsKey = true
		}
	}

	return fa
}

func loadPackage(path string) (*packages.Package, error) {
	cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedImports, Env: []string{"CGO_ENABLED=0"}}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	return pkgs[0], nil
}

func parseSourceType(sourceType string) (string, string) {
	idx := strings.LastIndexByte(sourceType, '.')
	if idx == -1 {
		return "", ""
	}

	return sourceType[0:idx], sourceType[idx+1:]
}

func render(data resource) error {
	tmplBuf := bytes.Buffer{}
	fmtBuf := bytes.Buffer{}
	importBuf := bytes.Buffer{}
	errBuf := bytes.Buffer{}
	filename := fmt.Sprintf("zz_%s_provider.go", data.Name)

	fmtC := exec.Command("gofmt")
	fmtC.Stdin = &tmplBuf
	fmtC.Stdout = &fmtBuf
	fmtC.Stderr = &errBuf

	importC := exec.Command("goimports")
	importC.Stdin = &fmtBuf
	importC.Stdout = &importBuf
	importC.Stderr = &errBuf

	if err := tmpls.Execute(&tmplBuf, data); err != nil {
		return err
	}

	if err := fmtC.Run(); err != nil {
		return errors.New(string(errBuf.Bytes()))
	}

	if err := importC.Run(); err != nil {
		return errors.New(string(errBuf.Bytes()))
	}

	dest, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, &importBuf)
	if err != nil {
		return err
	}

	println(filename)

	return nil
}
