package dubbo

import (
	"context"
	"fmt"

	"mosn.io/mosn/pkg/log"
	"mosn.io/mosn/pkg/protocol/xprotocol"
	"mosn.io/mosn/pkg/types"
)

/**
* Dubbo protocol
* Request & Response: (byte)
* 0           1           2           3           4           5           6           7           8
* +-----------+-----------+-----------+-----------+-----------+-----------+-----------+-----------+
* |magic high | magic low |  flag     | status    |               id                              |
* +-----------+-----------+-----------+-----------+-----------+-----------+-----------+-----------+
* |      id                                       |               data length                     |
* +-----------+-----------+-----------+-----------+-----------+-----------+-----------+-----------+
* |                               payload                                                         |
* +-----------------------------------------------------------------------------------------------+
* magic: 0xdabb
*
* flag: (bit offset)
* 0           1           2           3           4           5           6           7           8
* +-----------+-----------+-----------+-----------+-----------+-----------+-----------+-----------+
* |              serialization id                             |  event    | two way   |   req/rsp |
* +-----------+-----------+-----------+-----------+-----------+-----------+-----------+-----------+
* event: 1 mean ping
* two way: 1 mean req & rsp pair
* req/rsp: 1 mean req
 */
func init() {
	xprotocol.RegisterProtocol(ProtocolName, &dubboProtocol{})
}

var MagicTag = []byte{0xda, 0xbb}

type dubboProtocol struct{}

func (proto *dubboProtocol) Name() types.ProtocolName {
	return ProtocolName
}

func (proto *dubboProtocol) Encode(ctx context.Context, model interface{}) (types.IoBuffer, error) {
	if frame, ok := model.(*Frame); ok {
		if frame.Direction == EventRequest {
			return encodeRequest(ctx, frame)
		} else if frame.Direction == EventResponse {
			return encodeResponse(ctx, frame)
		}
	}
	log.Proxy.Errorf(ctx, "[protocol][dubbo] encode with unknown command : %+v", model)
	return nil, xprotocol.ErrUnknownType
}

func (proto *dubboProtocol) Decode(ctx context.Context, data types.IoBuffer) (interface{}, error) {
	if data.Len() >= HeaderLen {
		frame, err := decodeFrame(ctx, data)
		if err != nil {
			// unknown cmd type
			return nil, fmt.Errorf("[protocol][dubbo] Decode Error, type = %s , err = %v", UnKnownCmdType, err)
		}
		return frame, err
	}
	return nil, nil
}

// heartbeater
func (proto *dubboProtocol) Trigger(requestId uint64) xprotocol.XFrame {
	// not support
	return nil
}

func (proto *dubboProtocol) Reply(requestId uint64) xprotocol.XFrame {
	// TODO make readable
	return &Frame{
		Header: Header{
			Magic:   MagicTag,
			Flag:    0x22,
			Status:  0x14,
			Id:      requestId,
			DataLen: 0x02,
		},
		payload: []byte{0x4e, 0x4e},
	}
}

// hijacker
func (proto *dubboProtocol) Hijack(statusCode uint32) xprotocol.XFrame {
	// not support
	return nil
}

func (proto *dubboProtocol) Mapping(httpStatusCode uint32) uint32 {
	// not support
	return -1
}
