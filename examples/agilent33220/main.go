// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package main

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/gotmc/usbtmc"
	_ "github.com/gotmc/usbtmc/driver/google"
)

func main() {

	// Create new USBTMC context and new device.
	start := time.Now()
	ctx, err := usbtmc.NewContext()
	if err != nil {
		log.Fatalf("Error creating new USB context: %s", err)
	}
	fg, err := ctx.NewDevice("USB0::2391::1031::MY44035849::INSTR")
	if err != nil {
		log.Fatalf("NewDevice error: %s", err)
	}
	log.Printf("%.2fs to create new device.\n", time.Since(start).Seconds())

	// Configure function generator
	fg.WriteString("*CLS\n")
	fg.WriteString("burst:state off\n")
	fg.Write([]byte("apply:sinusoid 2340, 0.1, 0.0\n")) // Write using byte slice
	io.WriteString(fg, "burst:internal:period 0.112\n") // WriteString using io's Writer interface
	fg.WriteString("burst:internal:period 0.112\n")     // WriteString
	fg.WriteString("burst:ncycles 131\n")
	fg.WriteString("burst:state on\n")

	// Query using a write and then a read.
	queries := []string{"volt", "freq", "volt:offs", "volt:unit"}
	for _, q := range queries {
		ws := fmt.Sprintf("%s?\n", q)
		fg.WriteString(ws)
		var p [512]byte
		bytesRead, err := fg.Read(p[:])
		if err != nil {
			log.Printf("Error reading: %v", err)
		} else {
			log.Printf("Read %d bytes for %s? = %s\n", bytesRead, q, p)
		}
	}

	// Query using the query method
	for _, q := range queries {
		ws := fmt.Sprintf("%s?\n", q)
		s, err := fg.Query(ws)
		if err != nil {
			log.Printf("Error reading: %v", err)
		} else {
			log.Printf("Query %s? = %s\n", q, s)
		}
	}

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
