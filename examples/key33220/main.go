// Copyright (c) 2015-2020 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/gotmc/usbtmc"
	_ "github.com/gotmc/usbtmc/driver/google"
)

var (
	debugLevel uint
)

func init() {
	const (
		defaultLevel = 1
		debugUsage   = "USB debug level"
	)
	flag.UintVar(&debugLevel, "debug", defaultLevel, debugUsage)
	flag.UintVar(&debugLevel, "d", defaultLevel, debugUsage+" (shorthand)")
}

func main() {
	// Parse the config flags to determine the config JSON filename
	flag.Parse()

	// Create new USBTMC context and new device.
	start := time.Now()
	ctx, err := usbtmc.NewContext()
	if err != nil {
		log.Fatalf("Error creating new USB context: %s", err)
	}
	ctx.SetDebugLevel(int(debugLevel))

	fg, err := ctx.NewDevice("USB0::2391::1031::MY44035849::INSTR")
	if err != nil {
		log.Fatalf("NewDevice error: %s", err)
	}
	log.Printf("%.2fs to create new device.", time.Since(start).Seconds())

	// Configure function generator using different write methods.
	numCycles := 131
	period := 0.112
	fg.WriteString("*CLS\n")                              // Write using usbtmc.WriteString
	io.WriteString(fg, "burst:state off\n")               // Write using io.WriteString
	fg.Write([]byte("apply:sinusoid 2340, 0.1, 0.0\n"))   // Write using byte slice
	fmt.Fprintf(fg, "burst:internal:period %f\n", period) // Write using fmt.Fprint
	fg.Command("burst:ncycles %d", numCycles)             // Write using usbtmc.Command
	fg.Command("burst:state on")                          // Command appends a newline.

	queries := []string{"volt?", "freq?", "volt:offs?", "volt:unit?"}

	// Query using a write and then a read.
	for _, q := range queries {
		fg.WriteString(q)
		p := make([]byte, 512)
		bytesRead, err := fg.Read(p)
		if err != nil {
			log.Printf("Error reading: %v", err)
		} else {
			log.Printf("Read %d bytes for %s? = %s", bytesRead, q, p[:bytesRead])
		}
	}

	// Query using the query method
	queryRange(fg, queries)

	// Close the function generator and USBTMC context and check for errors.
	err = fg.Close()
	if err != nil {
		log.Printf("error closing fg: %s", err)
	}
	err = ctx.Close()
	if err != nil {
		log.Printf("Error closing context: %s", err)
	}
}

func queryRange(fg *usbtmc.Device, r []string) {
	for _, q := range r {
		s, err := fg.Query(q)
		if err != nil {
			log.Printf("Error reading: %v", err)
		} else {
			log.Printf("Query %s = %s", q, s)
		}
	}
}
