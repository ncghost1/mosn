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
	"context"
	"github.com/stretchr/testify/assert"
	"mosn.io/pkg/buffer"
	"runtime/debug"
	"testing"
)

func TestRespCodec(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("TestRespCodec error: %v %s", r, string(debug.Stack()))
		}
	}()

	testString := "$11\r\nhelloworld!\r\n"
	buf := buffer.NewIoBufferString(testString)
	if buf == nil {
		t.Error("New buffer failed.")
	}

	r := &Resp{
		Type:  TypeBulkBytes,
		Value: []byte("helloworld!"),
		Array: nil,
	}

	rc := RespCodec{}
	if ret, err := rc.Decode(context.Background(), buf); err != nil {
		t.Errorf("Decode failed:%v", err)
	} else {
		if resp, ok := ret.(*Resp); ok {
			assert.True(t, resp.IsBulkBytes() && string(resp.Value) == "helloworld!")
		} else {
			t.Errorf("Unexpected type")
		}
	}

	if ret, err := rc.Encode(context.Background(), r); err != nil {
		t.Errorf("Encode failed:%v", err)
	} else {
		assert.True(t, ret.String() == testString)
	}

	if p := rc.Name(); p != RESP {
		t.Errorf("get RespCodec Name failed，want:%v but got: %v", RESP, p)
	}
}
