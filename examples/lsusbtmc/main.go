package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/gotmc/usbtmc"
)

var (
	readonly = flag.Bool("readonly", false, "Only read from the controller")
	debug    = flag.Int("debug", 0, "USB debugging control")
)

func main() {
	flag.Parse()

	start := time.Now()
	ctx := usbtmc.NewContext()
	defer ctx.Close()

	if *debug != 0 {
		ctx.Debug(*debug)
	}

	start = time.Now()
	fg := ctx.NewInstrument("USB0::2391::1031::MY44027825::INSTR")
	log.Printf("%.2fs to setup instrument\n", time.Since(start).Seconds())
	start = time.Now()
	fmt.Printf(
		"Found the Arb Wave Gen S/N %s by Vendor ID %d with Product ID %d\n",
		fg.Device.Descriptor.SerialNumber,
		fg.Device.Descriptor.Vendor,
		fg.Device.Descriptor.Product)
	// Send commands to waveform generator
	fg.Write([]byte("apply:sinusoid 2340, 0.1, 0.0")) // Write using byte slice
	io.WriteString(fg, "burst:internal:period 0.112") // WriteString using io's Writer interface
	fg.WriteString("burst:internal:period 0.112")     // WriteString
	fg.WriteString("burst:ncycles 131")
	fg.WriteString("burst:state on")
	fg.WriteString("*idn?")

	start = time.Now()
	var buf [1024]byte
	bytesRead, err := fg.Read(buf[:])
	log.Printf("%.2fs to read %d bytes\n", time.Since(start).Seconds(), bytesRead)
	if err != nil {
		log.Printf("Error reading: %s", err)
	}
	fmt.Printf("Read %d bytes = %s", bytesRead, buf[12:bytesRead])
	fmt.Printf("Last rune read = %x\n", buf[bytesRead-1:bytesRead])
	fmt.Printf("Last rune read = %q\n", buf[bytesRead-1:bytesRead])
	fmt.Printf("Read %d bytes = %v\n", bytesRead, buf[:12])

	// log.Print(fg.Write("freq 2340"))
	// log.Print(scope.Ask("*idn?"))
	defer fg.Close()
	fmt.Printf("Goodbye arbitrary waveform generator %s\n", fg.Device.Descriptor.SerialNumber)

}
