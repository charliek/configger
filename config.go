package main

import (
	"github.com/BurntSushi/toml"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

type Configger struct {
	writers []*writer
}

func (c *Configger) WriteFiles() {
	for _, w := range c.writers {
		w.WriteTemplate()
	}
}

func (c *Configger) loop() {
	c.WriteFiles()

	ticker := time.NewTicker(time.Second * 1).C
	done := make(chan os.Signal, 1)
	reload := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, os.Kill, syscall.SIGTERM)
	signal.Notify(reload, syscall.SIGHUP)

	running := true
	for running {
		select {
		case <-ticker:
			c.WriteFiles()
		case <-reload:
			// TODO reload configuration
			log.Println("Reloading Configuration...")
		case <-done:
			log.Println("Exiting...")
			running = false
		}
	}
}

type Config struct {
	BasePath      string
	Address       string `toml:"address"`
	Src           string `toml:"src"`
	Dest          string `toml:"dest"`
	Owner         string `toml:"owner"`
	Group         string `toml:"group"`
	Mode          string `toml:"mode"`
	CheckCommand  string `toml:"check_cmd"`
	ReloadCommand string `toml:"reload_cmd"`
	RemoveWithKV  string `toml:"remove_with_kv"`
}

func defaultConfig() Config {
	return Config{
		Address: "127.0.0.1:8500",
		Owner:   "root",
		Group:   "root",
		Mode:    "0644",
	}
}

func loadConfigFile(basepath, path string) (Config, error) {
	conf := defaultConfig()
	conf.BasePath = basepath
	if !filepath.IsAbs(path) {
		path = filepath.Join(basepath, path)
	}
	log.Printf("Loading conf from %s", path)
	if _, err := toml.DecodeFile(path, &conf); err != nil {
		log.Printf("Error loading config file at path %s: %v", path)
		return conf, err
	}
	return conf, nil
}
