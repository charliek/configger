package main

import (
	"log"
	"os"
)

func main() {
	// TODO this should be pulled in from the command line.
	basepath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error determining base directory %v", err)
	}

	// TODO this should be pulled in from command line and load directory
	conf, err := loadConfigFile(basepath, "example.toml")
	if err != nil {
		log.Fatalf("Error loading config %v", err)
	}

	w := NewWriter(conf)

	configger := &Configger{
		writers: []*writer{w},
	}
	configger.loop()
}
