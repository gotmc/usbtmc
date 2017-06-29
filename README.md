# usbtmc
Go library to communicate with a USB Test and Measurement Class (USBTMC)
enabled USB device.

[![GoDoc][godoc image]][godoc link]
[![Build Status][travis image]][travis link]
[![License Badge][license image]][LICENSE.txt]

## Overview

[USBTMC][] is a USB device class specification for test equiment and
instrumentation devices, such as oscilloscopes, digital multimeters, and
function generators. USBTMC requires three endpoints:

- Control endpoint
- Bulk-OUT endpoint
- Bulk-IN endpoint

Additionally, the USBTMC subclass USB488 has an Interrupt-IN endpoint.

## USBTMC Descriptors

- Interface class = 0xFE (application-secific)
- Interface subclass = 0x03 (indicates USBTMC)

## Installation

```bash
$ go get github.com/gotmc/usbtmc
```

## Usage

To use the [usbtmc][gousbtmc] package, you must register which Go-based
[libusb][] interface library should be used.  [libusb][] is "a C library
that provides generic access to USB devices." There are three Go-based
libusb hardware interface libraries available:

- [github.com/truveris/gousb][truveris]
- [github.com/gotmc/libusb][golibusb] -- Not working currently
- [github.com/google/gousb][gousb] -- Not yet implemented

To indicate which libusb interface library should be used, include
**one** of the following blank imports:

```go
import _ "github.com/gotmc/usbtmc/driver/truveris"
import _ "github.com/gotmc/usbtmc/driver/libusb"
import _ "github.com/gotmc/usbtmc/driver/gousb"
```

## Documentation

Documentation can be found at either:

- <https://godoc.org/github.com/gotmc/usbtmc>
- <http://localhost:6060/pkg/github.com/gotmc/usbtmc/> after running `$
  godoc -http=:6060`

## Contributing

To contribute, please fork the repository, create a feature branch, and then
submit a [pull request][].

### Testing

Prior to submitting a [pull request][], please run:

```bash
$ make check
```

To update and view the test coverage report:

```bash
$ make cover
```

### Disclosure and Call for Help

While this package works, it does not fully implement the [USBTMC][]
specification.  Please submit pull requests as needed to increase
functionality, maintainability, or reliability.

## License

[usbtmc][gousbtmc] is released under the MIT license. Please see the
[LICENSE.txt][] file for more information.

[godoc image]: https://godoc.org/github.com/gotmc/usbtmc?status.svg
[godoc link]: https://godoc.org/github.com/gotmc/usbtmc
[golibusb]: https://github.com/gotmc/libusb
[gousb]: https://github.com/google/gousb
[libusb]: http://libusb.info
[LICENSE.txt]: https://github.com/gotmc/libusb/blob/master/LICENSE.txt
[license image]: https://img.shields.io/badge/license-MIT-blue.svg
[pull request]: https://help.github.com/articles/using-pull-requests
[Scott Chacon]: http://scottchacon.com/about.html
[travis image]: http://img.shields.io/travis/gotmc/usbtmc/master.svg
[travis link]: https://travis-ci.org/gotmc/usbtmc
[truveris]: https://github.com/truveris/gousb
[usbtmc]: http://www.usb.org/developers/docs/devclass_docs/
[gousbtmc]: https://github.com/gotmc/usbtmc
