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

package resp

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDecodeInvalidRequests(t *testing.T) {
	test := []string{
		"*hello\r\n",
		"*-100\r\n",
		"*3\r\nhi",
		"*3\r\nhi\r\n",
		"*4\r\n$1",
		"*4\r\n$1\r",
		"*4\r\n$1\n",
		"*2\r\n$3\r\nget\r\n$what?\r\nx\r\n",
		"*4\r\n$3\r\nget\r\n$1\r\nx\r\n",
		"*2\r\n$3\r\nget\r\n$1\r\nx",
		"*2\r\n$3\r\nget\r\n$1\r\nx\r",
		"*2\r\n$3\r\nget\r\n$100\r\nx\r\n",
		"$6\r\nfoobar\r",
		"$0\rn\r\n",
		"$-1\n",
		"*0",
		"*2n$3\r\nfoo\r\n$3\r\nbar\r\n",
		"*-\r\n",
		"+OK\n",
		"-Error message\r",
	}
	for _, s := range test {
		_, err := DecodeFromBytes([]byte(s))
		assert.Error(t, err)
	}
}

func TestDecodeSimpleRequest1(t *testing.T) {
	_, err := DecodeFromBytes([]byte("\r\n"))
	assert.Error(t, err)
}

func TestDecodeSimpleRequest2(t *testing.T) {
	test := []string{
		"hello world\r\n",
		"hello world    \r\n",
		"    hello world    \r\n",
		"    hello     world\r\n",
		"    hello     world    \r\n",
	}
	for _, s := range test {
		a, err := DecodeMultiBulkFromBytes([]byte(s))
		assert.NoError(t, err)
		assert.Equal(t, 2, len(a))
		assert.True(t, bytes.Equal(a[0].Value, []byte("hello")))
		assert.True(t, bytes.Equal(a[1].Value, []byte("world")))
	}
}

func TestDecodeSimpleRequest3(t *testing.T) {
	test := []string{"\r", "\n", " \n"}
	for _, s := range test {
		_, err := DecodeFromBytes([]byte(s))
		assert.Error(t, err)
	}
}

func TestDecodeBulkBytes(t *testing.T) {
	test := "*2\r\n$4\r\nLLEN\r\n$6\r\nmylist\r\n"
	resp, err := DecodeFromBytes([]byte(test))
	assert.NoError(t, err)
	assert.True(t, len(resp.Array) == 2)
	s1 := resp.Array[0]
	assert.True(t, bytes.Equal(s1.Value, []byte("LLEN")))
	s2 := resp.Array[1]
	assert.True(t, bytes.Equal(s2.Value, []byte("mylist")))
}

func TestDecoder(t *testing.T) {
	test := []string{
		"$6\r\nfoobar\r\n",
		"$0\r\n\r\n",
		"$-1\r\n",
		"*0\r\n",
		"*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n",
		"*3\r\n:1\r\n:2\r\n:3\r\n",
		"*-1\r\n",
		"+OK\r\n",
		"-Error message\r\n",
		"*2\r\n$1\r\n0\r\n*0\r\n",
		"*3\r\n$4\r\nEVAL\r\n$31\r\nreturn {1,2,{3,'Hello World!'}}\r\n$1\r\n0\r\n",
	}
	for _, s := range test {
		_, err := DecodeFromBytes([]byte(s))
		assert.NoError(t, err)
	}
}

type loopReader struct {
	buf []byte
	pos int
}

func (b *loopReader) Read(p []byte) (int, error) {
	if b.pos == len(b.buf) {
		b.pos = 0
	}
	n := copy(p, b.buf[b.pos:])
	b.pos += n
	return n, nil
}

func newBenchmarkDecoder(n int) (*Decoder, error) {
	r := NewArray([]*Resp{
		NewBulkBytes(make([]byte, n)),
	})
	p, err := EncodeToBytes(r)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	for i := 0; i < 128 && b.Len() < 1024*1024; i++ {
		_, err := b.Write(p)
		if err != nil {
			return nil, err
		}
	}
	return NewDecoderSize(&loopReader{buf: b.Bytes()}, 1024*128), nil
}

func benchmarkDecode(b *testing.B, n int) {
	d, err := newBenchmarkDecoder(n)
	assert.NoError(b, err)
	for i := 0; i < b.N; i++ {
		multi, err := d.DecodeMultiBulk()
		assert.NoError(b, err)
		assert.True(b, len(multi) == 1 && len(multi[0].Value) == n)
	}
}

func BenchmarkDecode16B(b *testing.B)  { benchmarkDecode(b, 16) }
func BenchmarkDecode64B(b *testing.B)  { benchmarkDecode(b, 64) }
func BenchmarkDecode512B(b *testing.B) { benchmarkDecode(b, 512) }
func BenchmarkDecode1K(b *testing.B)   { benchmarkDecode(b, 1024) }
func BenchmarkDecode2K(b *testing.B)   { benchmarkDecode(b, 1024*2) }
func BenchmarkDecode4K(b *testing.B)   { benchmarkDecode(b, 1024*4) }
func BenchmarkDecode16K(b *testing.B)  { benchmarkDecode(b, 1024*16) }
func BenchmarkDecode32K(b *testing.B)  { benchmarkDecode(b, 1024*32) }
func BenchmarkDecode128K(b *testing.B) { benchmarkDecode(b, 1024*128) }
