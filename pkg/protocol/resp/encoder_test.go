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
	"io"
	"strconv"
	"testing"
)

func TestEncodeString(t *testing.T) {
	resp := NewString([]byte("OK"))
	testEncodeAndCheck(t, resp, []byte("+OK\r\n"))
}

func TestEncodeError(t *testing.T) {
	resp := NewError([]byte("Error"))
	testEncodeAndCheck(t, resp, []byte("-Error\r\n"))
}

func TestEncodeInt(t *testing.T) {
	for _, v := range []int{-1, 0, 1024 * 1024} {
		s := strconv.Itoa(v)
		resp := NewInt([]byte(s))
		testEncodeAndCheck(t, resp, []byte(":"+s+"\r\n"))
	}
}

func TestEncodeBulkBytes(t *testing.T) {
	resp := NewBulkBytes(nil)
	testEncodeAndCheck(t, resp, []byte("$-1\r\n"))
	resp.Value = []byte{}
	testEncodeAndCheck(t, resp, []byte("$0\r\n\r\n"))
	resp.Value = []byte("helloworld!!")
	testEncodeAndCheck(t, resp, []byte("$12\r\nhelloworld!!\r\n"))
}

func TestEncodeArray(t *testing.T) {
	resp := NewArray(nil)
	testEncodeAndCheck(t, resp, []byte("*-1\r\n"))
	resp.Array = []*Resp{}
	testEncodeAndCheck(t, resp, []byte("*0\r\n"))
	resp.Array = append(resp.Array, NewInt([]byte(strconv.Itoa(0))))
	testEncodeAndCheck(t, resp, []byte("*1\r\n:0\r\n"))
	resp.Array = append(resp.Array, NewBulkBytes(nil))
	testEncodeAndCheck(t, resp, []byte("*2\r\n:0\r\n$-1\r\n"))
	resp.Array = append(resp.Array, NewBulkBytes([]byte("test")))
	testEncodeAndCheck(t, resp, []byte("*3\r\n:0\r\n$-1\r\n$4\r\ntest\r\n"))
}

func testEncodeAndCheck(t *testing.T, resp *Resp, expect []byte) {
	b, err := EncodeToBytes(resp)
	assert.NoError(t, err)
	assert.True(t, bytes.Equal(expect, b))
}

func newBenchmarkEncoder(n int) *Encoder {
	return NewEncoderSize(io.Discard, 1024*128)
}

func benchmarkEncode(b *testing.B, n int) {
	multi := []*Resp{
		NewBulkBytes(make([]byte, n)),
	}
	e := newBenchmarkEncoder(n)
	for i := 0; i < b.N; i++ {
		assert.NoError(b, e.EncodeMultiBulk(multi, false))
	}
	assert.NoError(b, e.Flush())
}

func BenchmarkEncode16B(b *testing.B)  { benchmarkEncode(b, 16) }
func BenchmarkEncode64B(b *testing.B)  { benchmarkEncode(b, 64) }
func BenchmarkEncode512B(b *testing.B) { benchmarkEncode(b, 512) }
func BenchmarkEncode1K(b *testing.B)   { benchmarkEncode(b, 1024) }
func BenchmarkEncode2K(b *testing.B)   { benchmarkEncode(b, 1024*2) }
func BenchmarkEncode4K(b *testing.B)   { benchmarkEncode(b, 1024*4) }
func BenchmarkEncode16K(b *testing.B)  { benchmarkEncode(b, 1024*16) }
func BenchmarkEncode32K(b *testing.B)  { benchmarkEncode(b, 1024*32) }
func BenchmarkEncode128K(b *testing.B) { benchmarkEncode(b, 1024*128) }
