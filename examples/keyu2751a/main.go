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

	// On power up, the U2751A shows PID 15896 (0x3E18), and I can't communicate
	// properly. After running the Agilent Measurement Manager, the PID shows
	// 15640 (0x3D18) and the U2751A works like any other USBTMC device.
	// I wasn't able to get close to working until the Agilent Measurement
	// Manager (AMM) udpated the firmware to 1.08. When the PID is correct, it
	// shows VISA address `USB0::0x0957::0x3D18::MY51010003::0::INSTR`
	sw, err := ctx.NewDeviceByVIDPID(2391, 15640)
	if err != nil {
		log.Fatalf("NewDevice error: %s", err)
	}
	log.Printf("%.2fs to create new device.", time.Since(start).Seconds())

	sw.WriteString("rout:clos (@302:304)")
	// time.Sleep(time.Second * 2)
	// sw.WriteString("rout:clos (@302:304)")

	log.Printf("Start querying")
	queries := []string{"*idn?", "rout:clos? (@301:305)"}

	// Query using the query method
	queryRange(sw, queries)

	sw.WriteString("rout:open (@302:304)")

	// Close the matrix switch and USBTMC context and check for errors.
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
