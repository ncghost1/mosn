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
	"errors"
	"fmt"
	"io"
	"mosn.io/mosn/pkg/protocol"
	"mosn.io/mosn/pkg/types"
	"mosn.io/pkg/buffer"
	"strconv"
)

type Encoder struct {
	bw  *protocol.BufWriter
	Err error
}

var ErrFailedEncoder = errors.New("use of failed encoder")

func NewEncoder(w io.Writer) *Encoder {
	return NewEncoderBuffer(protocol.NewBufWriterSize(w, 8192))
}

func NewEncoderSize(w io.Writer, size int) *Encoder {
	return NewEncoderBuffer(protocol.NewBufWriterSize(w, size))
}

func NewEncoderBuffer(bw *protocol.BufWriter) *Encoder {
	return &Encoder{bw: bw}
}

func (e *Encoder) Encode(r *Resp, flush bool) error {
	if e.Err != nil {
		return ErrFailedEncoder
	}
	if err := e.encodeResp(r); err != nil {
		e.Err = err
	} else if flush {
		e.Err = e.bw.Flush()
	}
	return e.Err
}

func (e *Encoder) EncodeMultiBulk(multi []*Resp, flush bool) error {
	if e.Err != nil {
		return ErrFailedEncoder
	}
	if err := e.encodeMultiBulk(multi); err != nil {
		e.Err = err
	} else if flush {
		e.Err = e.bw.Flush()
	}
	return e.Err
}

func (e *Encoder) Flush() error {
	if e.Err != nil {
		return ErrFailedEncoder
	}
	if err := e.bw.Flush(); err != nil {
		e.Err = err
	}
	return e.Err
}

func Encode(w io.Writer, r *Resp) error {
	return NewEncoder(w).Encode(r, true)
}

func EncodeToBytes(r *Resp) ([]byte, error) {
	var b = &bytes.Buffer{}
	if err := Encode(b, r); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func EncodeToIoBuffer(r *Resp) (types.IoBuffer, error) {
	b, err := EncodeToBytes(r)
	if err != nil {
		return nil, err
	}
	buf := buffer.GetIoBuffer(len(b))
	n, err := buf.Write(b)
	if err != nil {
		return nil, err
	}
	if n != len(b) {
		return nil, fmt.Errorf("short write! expected to write %d bytes, but actually wrote %d bytes", len(b), n)
	}
	return buf, nil
}

func (e *Encoder) encodeResp(r *Resp) error {
	if err := e.bw.WriteByte(byte(r.Type)); err != nil {
		return err
	}
	switch r.Type {
	default:
		return fmt.Errorf("bad resp type %v", r.Type)
	case TypeString, TypeError, TypeInt:
		return e.encodeTextBytes(r.Value)
	case TypeBulkBytes:
		return e.encodeBulkBytes(r.Value)
	case TypeArray:
		return e.encodeArray(r.Array)
	}
}

func (e *Encoder) encodeMultiBulk(multi []*Resp) error {
	if err := e.bw.WriteByte(byte(TypeArray)); err != nil {
		return err
	}
	return e.encodeArray(multi)
}

func (e *Encoder) encodeTextBytes(b []byte) error {
	if _, err := e.bw.Write(b); err != nil {
		return err
	}
	if _, err := e.bw.WriteString("\r\n"); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) encodeTextString(s string) error {
	if _, err := e.bw.WriteString(s); err != nil {
		return err
	}
	if _, err := e.bw.WriteString("\r\n"); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) encodeInt(v int64) error {
	return e.encodeTextString(strconv.FormatInt(v, 10))
}

func (e *Encoder) encodeBulkBytes(b []byte) error {
	if b == nil {
		return e.encodeInt(-1)
	} else {
		if err := e.encodeInt(int64(len(b))); err != nil {
			return err
		}
		return e.encodeTextBytes(b)
	}
}

func (e *Encoder) encodeArray(array []*Resp) error {
	if array == nil {
		return e.encodeInt(-1)
	} else {
		if err := e.encodeInt(int64(len(array))); err != nil {
			return err
		}
		for _, r := range array {
			if err := e.encodeResp(r); err != nil {
				return err
			}
		}
		return nil
	}
}
