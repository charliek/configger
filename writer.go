package main

import (
	"bytes"
	"crypto/md5"
	"github.com/armon/consul-api"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
)

type writer struct {
	config     Config
	lookup     *ConsulLookUp
	lastRunMd5 [16]byte
}

func NewWriter(c Config) *writer {
	client, _ := consulapi.NewClient(&consulapi.Config{
		Address:    c.Address,
		HttpClient: http.DefaultClient,
	})
	return &writer{
		config: c,
		lookup: &ConsulLookUp{
			client:       client,
			RemoveWithKV: c.RemoveWithKV,
		},
	}
}

func (w *writer) toFullPath(path string) string {
	if !filepath.IsAbs(path) {
		path = filepath.Join(w.config.BasePath, path)
	}
	return path
}

func (w *writer) loadTemplate() (*template.Template, error) {
	path := w.toFullPath(w.config.Src)
	funcMap := template.FuncMap{
		"lookupService": w.lookup.LookupService,
	}
	return template.New("conf").Funcs(funcMap).ParseFiles(path)
}

func (w *writer) reloadService() {
	if w.config.ReloadCommand != "" {
		// TODO need to implement service reloading
		log.Printf("Reloading service with command %s", w.config.ReloadCommand)
	}
}

func (w *writer) checkConfig([]byte) error {
	if w.config.CheckCommand != "" {
		// TODO need to implement configuration checking
		log.Printf("Checking config with command %s", w.config.CheckCommand)
	}
	return nil
}

func (w *writer) WriteTemplate() error {
	tmpl, err := w.loadTemplate()
	if err != nil {
		return err
	}
	path := w.toFullPath(w.config.Dest)

	ctx := map[string][]string{}
	var buff bytes.Buffer
	err = tmpl.ExecuteTemplate(&buff, "example.conf.tmpl", ctx)
	if err != nil {
		log.Printf("Error writing template to '%s': %v", path, err)
		return err
	}
	bits := buff.Bytes()
	checksum := md5.Sum(bits)
	if checksum != w.lastRunMd5 {
		w.lastRunMd5 = checksum
		err = w.checkConfig(bits)
		if err != nil {
			log.Printf("Not writing file %s. Error with dynamic temlate output %v", path, err)
			return err
		}

		ioutil.WriteFile(path, bits, w.config.FileMode)
		os.Chmod(path, w.config.FileMode)
		// TODO figure out how to chmod to proper user/group. Maybe shell out if needed.
		// os.Chown(path, t.Uid, t.Gid)
		log.Printf("Wrote out file %s", path)
		w.reloadService()
	}
	return nil
}
