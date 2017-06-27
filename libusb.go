// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import "github.com/gotmc/usbtmc/driver"

// Driver is the registered driver
var libusbDriver driver.Driver

// Register is called to register a driver for use by the program.
func Register(d driver.Driver) {
	libusbDriver = d
}
