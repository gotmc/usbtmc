// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package driver

// Driver defines the behavior required by types that want
// to implement a new search type.
type Driver interface {
	NewContext() (Context, error)
}

type Context interface {
	Close() error
	NewDeviceByVIDPID(VID, PID uint) (USBDevice, error)
	// NewDeviceBySerial(sn string) (USBDevice, error)
}

type USBDevice interface {
	Close() error
	String() string
	Write(p []byte) (n int, err error)
	WriteString(s string) (n int, err error)
	Read(p []byte) (n int, err error)
	Query(s string) (string, error)
}
