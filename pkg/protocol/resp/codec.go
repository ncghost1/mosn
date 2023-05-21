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
	"mosn.io/mosn/pkg/types"
)

type RespCodec struct {
}

func (c *RespCodec) Name() types.ProtocolName {
	return RESP
}

func (c *RespCodec) Encode(ctx context.Context, model interface{}) (types.IoBuffer, error) {
	r := model.(*Resp)
	buf, err := EncodeToIoBuffer(r) // Todo: Reuse encoder
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (c *RespCodec) Decode(ctx context.Context, data types.IoBuffer) (interface{}, error) {
	r, err := DecodeFromBytes(data.Bytes()) // Todo: Reuse decoder
	if err != nil {
		return nil, err
	}
	return r, nil
}
