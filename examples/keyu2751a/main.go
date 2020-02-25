// Copyright (c) 2015-2020 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package main

import (
	"flag"
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

	sw, err := ctx.NewDeviceByVIDPID(2391, 15896)
	if err != nil {
		log.Fatalf("NewDevice error: %s", err)
	}
	log.Printf("%.2fs to create new device.", time.Since(start).Seconds())

	// Configure switch matrix using different write methods.
	_, err = sw.WriteString("*CLS\n") // Write using usbtmc.WriteString
	if err != nil {
		log.Fatalf("error writing *CLS: %s", err)
	}

	log.Printf("Start querying")
	queries := []string{"syst:vers?"}

	// Query using a write and then a read.
	for _, q := range queries {
		sw.WriteString(q)
		p := make([]byte, 512)
		bytesRead, err := sw.Read(p)
		if err != nil {
			log.Printf("Error reading: %v", err)
		} else {
			log.Printf("Read %d bytes for %s? = %s", bytesRead, q, p[:bytesRead])
		}
	}

	// Query using the query method
	queryRange(sw, queries)

	// Close the function generator and USBTMC context and check for errors.
	err = sw.Close()
	if err != nil {
		log.Printf("error closing sw: %s", err)
	}
	err = ctx.Close()
	if err != nil {
		log.Printf("Error closing context: %s", err)
	}
}

func queryRange(sw *usbtmc.Device, r []string) {
	for _, q := range r {
		s, err := sw.Query(q)
		if err != nil {
			log.Printf("Error reading: %v", err)
		} else {
			log.Printf("Query %s = %s", q, s)
		}
	}
}
