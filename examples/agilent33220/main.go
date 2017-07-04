// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
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

func main() {
	flag.Parse()

	start := time.Now()
	ctx, err := usbtmc.NewContext()
	if err != nil {
		log.Fatalf("Error creating new USB context: %s", err)
	}

	start = time.Now()
	// fg, err := ctx.NewDeviceByVIDPID(0x957, 0x407)
	fg, err := ctx.NewDevice("USB0::2391::1031::MY44035849::INSTR")
	if err != nil {
		log.Fatalf("NewDevice error: %s", err)
	}
	log.Printf("%.2fs to setup instrument\n", time.Since(start).Seconds())
	start = time.Now()
	fg.WriteString("*CLS\n")
	fg.WriteString("burst:state off\n")
	fg.Write([]byte("apply:sinusoid 2340, 0.1, 0.0\n")) // Write using byte slice
	io.WriteString(fg, "burst:internal:period 0.112\n") // WriteString using io's Writer interface
	fg.WriteString("burst:internal:period 0.112\n")     // WriteString
	fg.WriteString("burst:ncycles 131\n")
	fg.WriteString("burst:state on\n")

	// Works to write *idn? to the fg and then read the response.
	log.Println("Start of *idn? write & read")
	fg.WriteString("*idn?\n")
	log.Println("Wrote *idn? to fgen.")
	start = time.Now()
	// TODO(mdr): Instead of using 512 magic number, I should read the maximum
	// buffer size for the bulk in endpoint from the USB device.
	var buf [512]byte
	log.Println("About to read 512 bytes from buffer.")
	bytesRead, err := fg.Read(buf[:])
	log.Printf("%.2fs to read %d bytes\n", time.Since(start).Seconds(), bytesRead)
	if err != nil {
		log.Printf("Error reading: %v", err)
	} else {
		log.Printf("Read %d bytes for \"*idn?\" = %s\n", bytesRead, buf)
	}

	// This works
	queries := []string{"volt", "freq", "volt:offs", "volt:unit"}
	for _, q := range queries {
		log.Printf("Start of %s? write & read.", q)
		ws := fmt.Sprintf("%s?\n", q)
		fg.WriteString(ws)
		var p [512]byte
		bytesRead, err = fg.Read(p[:])
		log.Printf("%.2fs to read %d bytes\n", time.Since(start).Seconds(), bytesRead)
		if err != nil {
			log.Printf("Error reading: %v", err)
		} else {
			log.Printf("Read %d bytes for %s? = %s\n", bytesRead, q, p)
		}
	}

	err = fg.Close()
	if err != nil {
		log.Printf("error closing fg: %s", err)
	}
	err = ctx.Close()
	if err != nil {
		log.Printf("Error closing context: %s", err)
	}
}
