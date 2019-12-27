// Copyright (c) 2015-2020 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package driver

// Driver defines the behavior required by types that want
// to implement a USBTMC driver.
type Driver interface {
	NewContext() (Context, error)
}

// Context defines the behavior required for USBTMC drivers.
type Context interface {
	Close() error
	SetDebugLevel(level int)
	NewDeviceByVIDPID(VID, PID uint) (USBDevice, error)
	// NewDeviceBySerial(sn string) (USBDevice, error)
}

// USBDevice defines the behavior for a USB device.
type USBDevice interface {
	Close() error
	String() string
	Write(p []byte) (n int, err error)
	WriteString(s string) (n int, err error)
	Read(p []byte) (n int, err error)
}
