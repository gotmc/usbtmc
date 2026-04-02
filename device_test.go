// Copyright (c) 2015-2026 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import (
	"context"
	"encoding/binary"
	"errors"
	"testing"
)

// mockUSBDevice records writes and replays reads for testing.
type mockUSBDevice struct {
	writes [][]byte // captured raw writes
	reads  [][]byte // queued responses to return from Read
	readN  int      // index into reads
	closed bool
}

func (m *mockUSBDevice) Write(p []byte) (int, error) {
	cp := make([]byte, len(p))
	copy(cp, p)
	m.writes = append(m.writes, cp)
	return len(p), nil
}

func (m *mockUSBDevice) WriteString(s string) (int, error) {
	return m.Write([]byte(s))
}

func (m *mockUSBDevice) Read(p []byte) (int, error) {
	if m.readN >= len(m.reads) {
		return 0, errors.New("mock: no more reads queued")
	}
	data := m.reads[m.readN]
	m.readN++
	n := copy(p, data)
	return n, nil
}

func (m *mockUSBDevice) Close() error {
	m.closed = true
	return nil
}

func (m *mockUSBDevice) String() string {
	return "mock"
}

// buildDevDepMsgInResponse builds a USBTMC DEV_DEP_MSG_IN response header
// with the given bTag and payload.
func buildDevDepMsgInResponse(bTag byte, payload []byte) []byte {
	hdr := make([]byte, usbtmcHeaderLen)
	hdr[0] = byte(devDepMsgIn)
	hdr[1] = bTag
	hdr[2] = invertbTag(bTag)
	hdr[3] = 0x00
	binary.LittleEndian.PutUint32(hdr[4:8], uint32(len(payload))) //nolint:gosec
	hdr[8] = 0x01                                                 // EOM
	resp := append(hdr, payload...)
	// Pad to 4-byte alignment.
	if m := len(resp) % 4; m != 0 {
		resp = append(resp, make([]byte, 4-m)...)
	}
	return resp
}

func newTestDevice(mock *mockUSBDevice) *Device {
	return &Device{
		usbDevice:       mock,
		bTag:            0,
		termChar:        '\n',
		termCharEnabled: true,
	}
}

func TestWriteSingleChunk(t *testing.T) {
	mock := &mockUSBDevice{}
	dev := newTestDevice(mock)

	data := []byte("*IDN?\n")
	n, err := dev.Write(data)
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if n != len(data) {
		t.Errorf("Write returned n=%d, want %d", n, len(data))
	}
	if len(mock.writes) != 1 {
		t.Fatalf("expected 1 USB write, got %d", len(mock.writes))
	}

	// Verify header: msgID=devDepMsgOut(1), bTag=1, transferSize=6, EOM=1.
	w := mock.writes[0]
	if w[0] != byte(devDepMsgOut) {
		t.Errorf("msgID = %d, want %d", w[0], devDepMsgOut)
	}
	if w[1] != 1 {
		t.Errorf("bTag = %d, want 1", w[1])
	}
	transferSize := binary.LittleEndian.Uint32(w[4:8])
	if transferSize != uint32(len(data)) { //nolint:gosec
		t.Errorf("transferSize = %d, want %d", transferSize, len(data))
	}
	if w[8] != 0x01 {
		t.Errorf("EOM = %d, want 1", w[8])
	}
	// Verify payload follows header.
	payload := w[bulkOutHeaderSize : bulkOutHeaderSize+len(data)]
	for i, b := range payload {
		if b != data[i] {
			t.Errorf("payload[%d] = %x, want %x", i, b, data[i])
		}
	}
}

func TestWriteMultiChunk(t *testing.T) {
	mock := &mockUSBDevice{}
	dev := newTestDevice(mock)

	// Create data larger than maxTransferSize - bulkOutHeaderSize (500 bytes).
	data := make([]byte, 600)
	for i := range data {
		data[i] = byte(i % 256)
	}
	n, err := dev.Write(data)
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if n != len(data) {
		t.Errorf("Write returned n=%d, want %d", n, len(data))
	}
	if len(mock.writes) != 2 {
		t.Fatalf("expected 2 USB writes, got %d", len(mock.writes))
	}

	// First chunk: EOM should be 0.
	if mock.writes[0][8] != 0x00 {
		t.Errorf("first chunk EOM = %d, want 0", mock.writes[0][8])
	}
	// Second chunk: EOM should be 1.
	if mock.writes[1][8] != 0x01 {
		t.Errorf("second chunk EOM = %d, want 1", mock.writes[1][8])
	}
}

func TestReadSingleTransfer(t *testing.T) {
	mock := &mockUSBDevice{}
	dev := newTestDevice(mock)

	payload := []byte("Keysight Technologies\n")
	// The first bTag after 0 will be 1.
	resp := buildDevDepMsgInResponse(1, payload)
	mock.reads = [][]byte{resp}

	buf := make([]byte, 100)
	n, err := dev.Read(buf)
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if n != len(payload) {
		t.Errorf("Read returned n=%d, want %d", n, len(payload))
	}
	if string(buf[:n]) != string(payload) {
		t.Errorf("Read data = %q, want %q", buf[:n], payload)
	}

	// Verify the request header was sent (requestDevDepMsgIn).
	if len(mock.writes) != 1 {
		t.Fatalf("expected 1 write for request header, got %d", len(mock.writes))
	}
	if mock.writes[0][0] != byte(requestDevDepMsgIn) {
		t.Errorf("request msgID = %d, want %d", mock.writes[0][0], requestDevDepMsgIn)
	}
}

func TestBulkReadNoTermChar(t *testing.T) {
	mock := &mockUSBDevice{}
	dev := newTestDevice(mock)

	payload := []byte{0x01, 0x02, 0x03, 0x04}
	resp := buildDevDepMsgInResponse(1, payload)
	mock.reads = [][]byte{resp}

	buf := make([]byte, 100)
	n, err := dev.BulkRead(buf)
	if err != nil {
		t.Fatalf("BulkRead returned error: %v", err)
	}
	if n != len(payload) {
		t.Errorf("BulkRead returned n=%d, want %d", n, len(payload))
	}

	// Verify termCharEnabled is NOT set in the request header (bit 1 of byte 8).
	reqHeader := mock.writes[0]
	if reqHeader[8]&0x02 != 0 {
		t.Error("BulkRead request has termCharEnabled set, want unset")
	}
}

func TestCommand(t *testing.T) {
	mock := &mockUSBDevice{}
	dev := newTestDevice(mock)

	err := dev.Command(context.Background(), "FREQ %d", 1000)
	if err != nil {
		t.Fatalf("Command returned error: %v", err)
	}
	if len(mock.writes) != 1 {
		t.Fatalf("expected 1 USB write, got %d", len(mock.writes))
	}

	// Extract payload from the write (skip 12-byte header).
	w := mock.writes[0]
	transferSize := binary.LittleEndian.Uint32(w[4:8])
	payload := string(w[bulkOutHeaderSize : bulkOutHeaderSize+transferSize])
	expected := "FREQ 1000\n"
	if payload != expected {
		t.Errorf("Command payload = %q, want %q", payload, expected)
	}
}

func TestQuery(t *testing.T) {
	mock := &mockUSBDevice{}
	dev := newTestDevice(mock)

	respPayload := []byte("1.00000E+03\n")
	// Query does a Write (bTag becomes 1), then a Read (bTag becomes 2).
	resp := buildDevDepMsgInResponse(2, respPayload)
	mock.reads = [][]byte{resp}

	result, err := dev.Query(context.Background(), "*IDN?")
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	if result != string(respPayload) {
		t.Errorf("Query result = %q, want %q", result, respPayload)
	}

	// Should have 2 writes: one for Command, one for Read request header.
	if len(mock.writes) != 2 {
		t.Fatalf("expected 2 USB writes, got %d", len(mock.writes))
	}
}

func TestWriteContextCancellation(t *testing.T) {
	mock := &mockUSBDevice{}
	dev := newTestDevice(mock)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	_, err := dev.WriteContext(ctx, []byte("data"))
	if err == nil {
		t.Fatal("WriteContext with cancelled context should return error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
	if len(mock.writes) != 0 {
		t.Errorf("expected 0 USB writes, got %d", len(mock.writes))
	}
}

func TestReadContextCancellation(t *testing.T) {
	mock := &mockUSBDevice{}
	dev := newTestDevice(mock)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	buf := make([]byte, 100)
	_, err := dev.ReadContext(ctx, buf)
	if err == nil {
		t.Fatal("ReadContext with cancelled context should return error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}

func TestReadBTagMismatch(t *testing.T) {
	mock := &mockUSBDevice{}
	dev := newTestDevice(mock)

	// Build response with wrong bTag (99 instead of expected 1).
	payload := []byte("data")
	resp := buildDevDepMsgInResponse(99, payload)
	mock.reads = [][]byte{resp}

	buf := make([]byte, 100)
	_, err := dev.Read(buf)
	if err == nil {
		t.Fatal("Read with mismatched bTag should return error")
	}
	if !contains(err.Error(), "bTag mismatch") {
		t.Errorf("error = %v, want bTag mismatch error", err)
	}
}

func TestReadWrongMsgID(t *testing.T) {
	mock := &mockUSBDevice{}
	dev := newTestDevice(mock)

	// Build a response with wrong msgID (devDepMsgOut instead of devDepMsgIn).
	resp := buildDevDepMsgInResponse(1, []byte("data"))
	resp[0] = byte(devDepMsgOut)
	mock.reads = [][]byte{resp}

	buf := make([]byte, 100)
	_, err := dev.Read(buf)
	if err == nil {
		t.Fatal("Read with wrong MsgID should return error")
	}
	if !contains(err.Error(), "unexpected MsgID") {
		t.Errorf("error = %v, want unexpected MsgID error", err)
	}
}

func TestClose(t *testing.T) {
	mock := &mockUSBDevice{}
	dev := newTestDevice(mock)

	err := dev.Close()
	if err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	if !mock.closed {
		t.Error("expected underlying USB device to be closed")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
