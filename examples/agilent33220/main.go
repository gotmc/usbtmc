// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package main

import (
	"flag"
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
	defer ctx.Close()

	start = time.Now()
	// fg, err := ctx.NewDeviceByVIDPID(0x957, 0x407)
	fg, err := ctx.NewDevice("USB0::2391::1031::MY44035849::INSTR")
	if err != nil {
		log.Fatalf("NewDevice error: %s", err)
	}
	log.Printf("%.2fs to setup instrument\n", time.Since(start).Seconds())
	start = time.Now()
	// fmt.Printf(
	// "Found the Arb Wave Gen S/N %s by Vendor ID %d with Product ID %d\n",
	// fg.USBDevice.Descriptor.SerialNumber,
	// fg.USBDevice.Descriptor.Vendor,
	// fg.USBDevice.Descriptor.Product)
	// Send commands to waveform generator
	fg.Write([]byte("apply:sinusoid 2340, 0.1, 0.0\n")) // Write using byte slice
	io.WriteString(fg, "burst:internal:period 0.112\n") // WriteString using io's Writer interface
	fg.WriteString("burst:internal:period 0.112\n")     // WriteString
	fg.WriteString("burst:ncycles 131\n")
	fg.WriteString("burst:state on\n")

	// Works to write *idn? to the fg and then read the response.
	fg.WriteString("*idn?\n")
	start = time.Now()
	var buf [512]byte
	bytesRead, err := fg.Read(buf[:])
	log.Printf("%.2fs to read %d bytes\n", time.Since(start).Seconds(), bytesRead)
	if err != nil {
		log.Printf("Error reading: %v", err)
	} else {
		log.Printf("Read %d bytes for \"*idn?\" = %s\n", bytesRead, buf)
	}

	// This works
	fg.WriteString("VOLT?\n")
	var volts [512]byte
	bytesRead, err = fg.Read(volts[:])
	log.Printf("%.2fs to read %d bytes\n", time.Since(start).Seconds(), bytesRead)
	if err != nil {
		log.Printf("Error reading: %v", err)
	} else {
		log.Printf("Read %d bytes for \"VOLT?\" = %s\n", bytesRead, volts)
	}

	log.Println("Not yet at bufio.NewScanner")
	// This works
	// fg.WriteString("FREQ?\n")
	// scanner := bufio.NewScanner(fg)
	// for scanner.Scan() {
	// log.Printf("WriteString(\"FREQ?\\n\") scanner.Text result: %s", scanner.Text())
	// }
	// log.Println("I'm here.")

	// result, err := fg.Query("FREQ?\n")
	// if err != nil {
	// log.Printf("query error: %s", err)
	// }
	// log.Printf("Query result: %s", result)
	fg.Close()

}
