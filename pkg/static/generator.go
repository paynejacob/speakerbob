//+build ignore
//

package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func fmtByteSlice(s []byte) string {
	builder := strings.Builder{}

	for _, v := range s {
		builder.WriteString(fmt.Sprintf("%d,", int(v)))
	}

	return builder.String()
}

var conv = map[string]interface{}{"conv": fmtByteSlice}
var tmpl = template.Must(template.New("").Funcs(conv).Parse(`
// Code generated by generator.go DO NOT EDIT

package static

var StaticMap = map[string][]byte{
{{- range $name, $file := . }}
	"{{ $name }}": []byte{ {{ conv $file }} },
{{- end }}
}

`))

const assetDir = "../../assets/"
const target = "zz_static_map.go"

type queue struct {
	q []string
}

func newQueue() *queue {
	return &queue{
		q: make([]string, 0),
	}
}

func (q *queue) empty() bool {
	return len(q.q) == 0
}

func (q *queue) push(v string) {
	q.q = append(q.q, v)
}

func (q *queue) pop() string {
	r := q.q[0]

	q.q = q.q[1:]

	return r
}

func main() {
	var err error

	if _, err = os.Stat(assetDir); os.IsNotExist(err) {
		log.Fatal("assets directory does not exist. Did you run scripts/build-ui?")
	}

	files := map[string][]byte{}
	pathQueue := newQueue()
	pathQueue.push(assetDir)

	for {
		if pathQueue.empty() {
			break
		}

		root := pathQueue.pop()

		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			relativePath := filepath.ToSlash(strings.TrimPrefix(path, assetDir))

			if info.IsDir() {
				if root == path {
					return nil
				}
				pathQueue.push(path)
				return nil
			}

			var raw []byte
			var compressed bytes.Buffer
			raw, err = ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			gz := gzip.NewWriter(&compressed)

			_, err = gz.Write(raw)
			if err != nil {
				log.Fatal("failed to compress file")
			}

			if err = gz.Flush(); err != nil {
				log.Fatal("failed to compress file")
			}

			if err = gz.Close(); err != nil {
				log.Fatal("failed to compress file")
			}

			files[relativePath] = compressed.Bytes()

			return nil
		})

		if err != nil {
			log.Fatal("error walking assets folder", err)
		}
	}

	buf := &bytes.Buffer{}
	if _, err := os.Create(target); err != nil {
		log.Fatal("error creating target file", err)
	}

	if err := tmpl.Execute(buf, files); err != nil {
		log.Fatal("error executing template", err)
	}

	out, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal("error formatting code", err)
	}

	if err = ioutil.WriteFile(target, out, os.ModePerm); err != nil {
		log.Fatal("error writing output file", err)
	}
}