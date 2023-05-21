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
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

func TestIDGenerator(t *testing.T) {
	if n := GenerateID(); n != 1 {
		t.Error("GenerateID failed.")
	}

	if n := GenerateIDString(); n != "2" {
		t.Error("GenerateIDString failed.")
	}

	if n := StreamIDConv(1); n != "1" {
		t.Error("StreamIDConv failed.")
	}
	if n := RequestIDConv("1"); n != 1 {
		t.Error("RequestIDConv failed.")
	}
}

func newReader(n int, input string) *BufReader {
	return &BufReader{rd: strings.NewReader(input), buf: make([]byte, n)}
}

func TestRead(t *testing.T) {
	var b bytes.Buffer
	for i := 0; i < 10; i++ {
		fmt.Fprintf(&b, "hello world %d", i)
	}
	var input = b.String()
	for n := 1; n < len(input); n++ {
		r := newReader(n, input)
		b := make([]byte, len(input))
		_, err := io.ReadFull(r, b)
		assert.NoError(t, err)
		assert.Equal(t, input, string(b))
	}
}

func TestReadByte(t *testing.T) {
	var input = "hello world"
	for n := 1; n < len(input); n++ {
		r := newReader(n, input)
		var s string
		for i := 0; i < len(input); i++ {
			b1, err := r.PeekByte()
			assert.NoError(t, err)
			b2, err := r.ReadByte()
			assert.NoError(t, err)
			assert.Equal(t, b1, b2)
			s += string(b1)
		}
		assert.Equal(t, input, s)
	}
}

func TestReadBytes(t *testing.T) {
	var b bytes.Buffer
	for i := 0; i < 10; i++ {
		fmt.Fprintf(&b, "hello world %d ", i)
	}
	var input = b.String()
	for n := 1; n < len(input); n++ {
		r := newReader(n, input)
		var s string
		for i := 0; i < 30; i++ {
			b, err := r.ReadBytes(' ')
			assert.NoError(t, err)
			s += string(b)
		}
		assert.Equal(t, input, s)
	}
}

func TestReadFull(t *testing.T) {
	var b bytes.Buffer
	for i := 0; i < 10; i++ {
		fmt.Fprintf(&b, "hello world %d ", i)
	}
	var input = b.String()
	for n := 1; n < len(input); n++ {
		r := newReader(n, input)
		b, err := r.ReadFull(len(input))
		assert.NoError(t, err)
		assert.Equal(t, string(b), input)
	}
}

func newBufWriter(n int, b *bytes.Buffer) *BufWriter {
	return &BufWriter{wr: b, buf: make([]byte, n)}
}

func TestWrite(t *testing.T) {
	for n := 1; n < 20; n++ {
		var input string
		var b bytes.Buffer
		var w = newBufWriter(n, &b)
		for i := 0; i < 10; i++ {
			s := fmt.Sprintf("hello world %d", i)
			_, err := w.Write([]byte(s))
			assert.NoError(t, err)
			input += s
		}
		assert.NoError(t, w.Flush())
		assert.Equal(t, b.String(), input)
	}
}

func TestWriteBytes(t *testing.T) {
	var input = "hello world!!"
	for n := 1; n < 20; n++ {
		var b bytes.Buffer
		var w = newBufWriter(n, &b)
		for i := 0; i < len(input); i++ {
			assert.NoError(t, w.WriteByte(input[i]))
		}
		assert.NoError(t, w.Flush())
		assert.Equal(t, b.String(), input)
	}
}

func TestWriteString(t *testing.T) {
	for n := 1; n < 20; n++ {
		var input string
		var b bytes.Buffer
		var w = newBufWriter(n, &b)
		for i := 0; i < 10; i++ {
			s := fmt.Sprintf("hello world %d", i)
			_, err := w.WriteString(s)
			assert.NoError(t, err)
			input += s
		}
		assert.NoError(t, w.Flush())
		assert.Equal(t, b.String(), input)
	}
}
