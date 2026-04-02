// Copyright (c) 2015-2026 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

// Package usbtmc provides a USB Test & Measurement Class (USBTMC) for
// interface for controlling test equipment over Ethernet ports using SCPI
// commands. It implements the VISA LXI resource string format and serves as an
// instrument driver for the ivi and visa packages.
//
// This package is part of the gotmc ecosystem. The visa package
// (github.com/gotmc/visa) defines a common interface for instrument
// communication across different transports (GPIB, USB, TCP/IP, serial). The
// usbtmc package provides the USB transport implementation. The ivi package
// (github.com/gotmc/ivi) builds on top of visa to provide standardized,
// instrument-class-specific APIs following the IVI Foundation specifications.
//
// Devices are addressed using VISA resource strings of the form:
//
//	USB<boardIndex>::<manufactuerID>::<modelCode>::<serialNumber>::INSTR
//
// For example:
//
//	USB0::2391::1031::MY44035849::INSTR
package usbtmc
