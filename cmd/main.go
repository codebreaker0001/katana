package main

import (
	log "github.com/sirupsen/logrus"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.WithError(err).Fatalln("cannot start cmd")
	}
}
