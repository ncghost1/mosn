/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package protocol

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"sync/atomic"
)

var defaultGenerator IDGenerator

// IDGenerator utility to generate auto-increment ids
type IDGenerator struct {
	counter uint64
}

// Get get id
func (g *IDGenerator) Get() uint64 {
	return atomic.AddUint64(&g.counter, 1)
}

// Get get id in string format
func (g *IDGenerator) GetString() string {
	n := atomic.AddUint64(&g.counter, 1)
	return strconv.FormatUint(n, 10)
}

// GenerateID get id by default global generator
func GenerateID() uint64 {
	return defaultGenerator.Get()
}

// GenerateIDString get id string by default global generator
func GenerateIDString() string {
	return defaultGenerator.GetString()
}

// StreamIDConv convert streamID from uint64 to string
func StreamIDConv(streamID uint64) string {
	return strconv.FormatUint(streamID, 10)
}

// RequestIDConv convert streamID from string to uint64
func RequestIDConv(streamID string) uint64 {
	reqID, _ := strconv.ParseUint(streamID, 10, 64)
	return reqID
}

const DefaultBufferSize = 1024

type BufReader struct {
	err error
	buf []byte

	rd   io.Reader
	rpos int
	wpos int

	slice sliceAlloc
}

type sliceAlloc struct {
	buf []byte
}

func (d *sliceAlloc) Make(n int) (ss []byte) {
	switch {
	case n == 0:
		return []byte{}
	case n >= 512:
		return make([]byte, n)
	default:
		if len(d.buf) < n {
			d.buf = make([]byte, 8192)
		}
		ss, d.buf = d.buf[:n:n], d.buf[n:]
		return ss
	}
}

func NewBufReader(rd io.Reader) *BufReader {
	return NewBufReaderSize(rd, DefaultBufferSize)
}

func NewBufReaderSize(rd io.Reader, size int) *BufReader {
	if size <= 0 {
		size = DefaultBufferSize
	}
	return &BufReader{rd: rd, buf: make([]byte, size)}
}

func NewBufReaderBuffer(rd io.Reader, buf []byte) *BufReader {
	if len(buf) == 0 {
		buf = make([]byte, DefaultBufferSize)
	}
	return &BufReader{rd: rd, buf: buf}
}

func (b *BufReader) fill() error {
	if b.err != nil {
		return b.err
	}
	if b.rpos > 0 {
		n := copy(b.buf, b.buf[b.rpos:b.wpos])
		b.rpos = 0
		b.wpos = n
	}
	n, err := b.rd.Read(b.buf[b.wpos:])
	if err != nil {
		b.err = err
	} else if n == 0 {
		b.err = io.ErrNoProgress
	} else {
		b.wpos += n
	}
	return b.err
}

func (b *BufReader) buffered() int {
	return b.wpos - b.rpos
}

func (b *BufReader) Read(p []byte) (int, error) {
	if b.err != nil || len(p) == 0 {
		return 0, b.err
	}
	if b.buffered() == 0 {
		if len(p) >= len(b.buf) {
			n, err := b.rd.Read(p)
			if err != nil {
				b.err = err
			}
			return n, b.err
		}
		if b.fill() != nil {
			return 0, b.err
		}
	}
	n := copy(p, b.buf[b.rpos:b.wpos])
	b.rpos += n
	return n, nil
}

func (b *BufReader) ReadByte() (byte, error) {
	if b.err != nil {
		return 0, b.err
	}
	if b.buffered() == 0 {
		if b.fill() != nil {
			return 0, b.err
		}
	}
	c := b.buf[b.rpos]
	b.rpos += 1
	return c, nil
}

func (b *BufReader) PeekByte() (byte, error) {
	if b.err != nil {
		return 0, b.err
	}
	if b.buffered() == 0 {
		if b.fill() != nil {
			return 0, b.err
		}
	}
	c := b.buf[b.rpos]
	return c, nil
}

func (b *BufReader) ReadSlice(delim byte) ([]byte, error) {
	if b.err != nil {
		return nil, b.err
	}
	for {
		var index = bytes.IndexByte(b.buf[b.rpos:b.wpos], delim)
		if index >= 0 {
			limit := b.rpos + index + 1
			slice := b.buf[b.rpos:limit]
			b.rpos = limit
			return slice, nil
		}
		if b.buffered() == len(b.buf) {
			b.rpos = b.wpos
			return b.buf, bufio.ErrBufferFull
		}
		if b.fill() != nil {
			return nil, b.err
		}
	}
}

func (b *BufReader) ReadBytes(delim byte) ([]byte, error) {
	var full [][]byte
	var last []byte
	var size int
	for last == nil {
		f, err := b.ReadSlice(delim)
		if err != nil {
			if err != bufio.ErrBufferFull {
				return nil, b.err
			}
			dup := b.slice.Make(len(f))
			copy(dup, f)
			full = append(full, dup)
		} else {
			last = f
		}
		size += len(f)
	}
	var n int
	var buf = b.slice.Make(size)
	for _, frag := range full {
		n += copy(buf[n:], frag)
	}
	copy(buf[n:], last)
	return buf, nil
}

func (b *BufReader) ReadFull(n int) ([]byte, error) {
	if b.err != nil || n == 0 {
		return nil, b.err
	}
	var buf = b.slice.Make(n)
	if _, err := io.ReadFull(b, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

type BufWriter struct {
	err error
	buf []byte

	wr   io.Writer
	wpos int
}

func NewBufWriter(wr io.Writer) *BufWriter {
	return NewBufWriterSize(wr, DefaultBufferSize)
}

func NewBufWriterSize(wr io.Writer, size int) *BufWriter {
	if size <= 0 {
		size = DefaultBufferSize
	}
	return &BufWriter{wr: wr, buf: make([]byte, size)}
}

func NewBufWriterBuffer(wr io.Writer, buf []byte) *BufWriter {
	if len(buf) == 0 {
		buf = make([]byte, DefaultBufferSize)
	}
	return &BufWriter{wr: wr, buf: buf}
}

func (b *BufWriter) Flush() error {
	return b.flush()
}

func (b *BufWriter) flush() error {
	if b.err != nil {
		return b.err
	}
	if b.wpos == 0 {
		return nil
	}
	n, err := b.wr.Write(b.buf[:b.wpos])
	if err != nil {
		b.err = err
	} else if n < b.wpos {
		b.err = io.ErrShortWrite
	} else {
		b.wpos = 0
	}
	return b.err
}

func (b *BufWriter) available() int {
	return len(b.buf) - b.wpos
}

func (b *BufWriter) Write(p []byte) (nn int, err error) {
	for b.err == nil && len(p) > b.available() {
		var n int
		if b.wpos == 0 {
			n, b.err = b.wr.Write(p)
		} else {
			n = copy(b.buf[b.wpos:], p)
			b.wpos += n
			b.flush()
		}
		nn, p = nn+n, p[n:]
	}
	if b.err != nil || len(p) == 0 {
		return nn, b.err
	}
	n := copy(b.buf[b.wpos:], p)
	b.wpos += n
	return nn + n, nil
}

func (b *BufWriter) WriteByte(c byte) error {
	if b.err != nil {
		return b.err
	}
	if b.available() == 0 && b.flush() != nil {
		return b.err
	}
	b.buf[b.wpos] = c
	b.wpos += 1
	return nil
}

func (b *BufWriter) WriteString(s string) (nn int, err error) {
	for b.err == nil && len(s) > b.available() {
		n := copy(b.buf[b.wpos:], s)
		b.wpos += n
		b.flush()
		nn, s = nn+n, s[n:]
	}
	if b.err != nil || len(s) == 0 {
		return nn, b.err
	}
	n := copy(b.buf[b.wpos:], s)
	b.wpos += n
	return nn + n, nil
}
