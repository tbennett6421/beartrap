/*
 * Copyright (c) 2015, Chris Benedict <chrisbdaemon@gmail.com>
 * All rights reserved.
 *
 * Licensing terms are located in LICENSE file.
 */

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chrisbdaemon/beartrap/alert"
	"github.com/chrisbdaemon/beartrap/broadcast"
	"github.com/chrisbdaemon/beartrap/config"
	"github.com/chrisbdaemon/beartrap/trap"
	getopt "github.com/kesselborn/go-getopt"
)

func main() {
	options := getOptions()

	cfg, err := config.New(options["config"].String)
	if err != nil {
		log.Fatal(err)
	}

	trapParams, err := cfg.TrapParams()
	if err != nil {
		log.Fatalf("Error reading traps: %s", err)
	}

	var broadcast broadcast.Broadcast

	// Hack as fill-in for handler
	c := make(chan alert.Alert)
	go func(c chan alert.Alert) {
		for {
			a := <-c
			log.Println(a.Message)
		}
	}(c)
	broadcast.AddReceiver(c)

	// Create and validate traps
	traps, err := initTraps(trapParams, broadcast)
	if err != nil {
		log.Fatalf("Error initializing traps: %s", err)
	}
	errors := validateTraps(traps)

	// If validation failed, report and quit,
	// if not, turn traps on
	if len(errors) > 0 {
		displayErrors(errors)
		os.Exit(-1)
	} else {
		startTraps(traps)

		// Hack to let traps run till I create a better mainloop
		for {
			time.Sleep(500 * time.Second)
		}
	}
}

func validateTraps(traps []trap.Interface) []error {
	var errors []error
	for i := range traps {
		errors = append(errors, traps[i].Validate()...)
	}
	return errors
}

// displayErrors takes a slice of errors and prints them to the screen
func displayErrors(errors []error) {
	for i := range errors {
		log.Println(errors[i])
	}
}

// startTraps takes a slice of traps and starts them in a goroutine
// TODO: Allow them to be stopped
func startTraps(traps []trap.Interface) {
	for i := range traps {
		go traps[i].Start()
	}
}

// initTraps take in a list of trap parameters, creates trap objects
// that are returned along with any errors generated from validation
func initTraps(trapParams []config.Params, d broadcast.Broadcast) ([]trap.Interface, error) {
	traps := []trap.Interface{}

	for i := range trapParams {
		trap, err := trap.New(trapParams[i], d)
		if err != nil {
			return nil, err
		}

		traps = append(traps, trap)
	}

	return traps, nil
}

// Parse commandline arguments into getopt object
func getOptions() map[string]getopt.OptionValue {
	optionDefinition := getopt.Options{
		Description: "Beartrap v0.3 by Chris Benedict <chrisbdaemon@gmail.com>",
		Definitions: getopt.Definitions{
			{"config|c|BEARTRAP_CONFIG", "configuration file", getopt.Required, ""},
		},
	}

	options, _, _, err := optionDefinition.ParseCommandLine()

	help, wantsHelp := options["help"]

	if err != nil || wantsHelp {
		exitCode := 0

		switch {
		case wantsHelp && help.String == "usage":
			fmt.Print(optionDefinition.Usage())
		case wantsHelp && help.String == "help":
			fmt.Print(optionDefinition.Help())
		default:
			fmt.Println("**** Error: ", err.Error(), "\n", optionDefinition.Help())
			exitCode = err.ErrorCode
		}
		os.Exit(exitCode)
	}

	return options
}
