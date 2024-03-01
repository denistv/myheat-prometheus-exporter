package main

import "github.com/denistv/wdlogger/wrappers/stdwrap"

func main() {
	logger := stdwrap.NewSTDWrapper()

	logger.Info("exporter started")
}
