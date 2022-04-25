/*
 * Copyright 2021 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package protobuf

import (
	"context"
	"encoding/binary"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/cloudwego/kitex/pkg/protocol/bprotoc"
	"github.com/cloudwego/kitex/pkg/remote"
)

// gogoproto generate
type marshaler interface {
	MarshalTo(data []byte) (n int, err error)
	Size() int
}

type protobufV2MsgCodec interface {
	XXX_Unmarshal(b []byte) error
	XXX_Marshal(b []byte, deterministic bool) ([]byte, error)
}

type grpcCodec struct{}

// NewGRPCCodec create grpc and protobuf codec
func NewGRPCCodec() remote.Codec {
	return new(grpcCodec)
}

func (c *grpcCodec) Encode(ctx context.Context, message remote.Message, out remote.ByteBuffer) (err error) {
	writer, ok := out.(remote.FrameWrite)
	if !ok {
		return fmt.Errorf("output buffer must implement FrameWrite")
	}
	var header [5]byte
	var data []byte
	switch t := message.Data().(type) {
	case bprotoc.FastWrite:
		// TODO: reuse data buffer when we can free it safely
		data = make([]byte, t.Size())
		t.FastWrite(data)
		fmt.Printf("Encode: (data=%x)\n", data[:9])
	case marshaler:
		// TODO: reuse data buffer when we can free it safely
		data = make([]byte, t.Size())
		if _, err = t.MarshalTo(data); err != nil {
			return err
		}
	case protobufV2MsgCodec:
		data, err = t.XXX_Marshal(nil, true)
		if err != nil {
			return err
		}
	case proto.Message:
		data, err = proto.Marshal(t)
		if err != nil {
			return err
		}
	case protobufMsgCodec:
		data, err = t.Marshal(nil)
		if err != nil {
			return err
		}
	}
	if err = writer.WriteData(data); err != nil {
		return err
	}
	binary.BigEndian.PutUint32(header[1:5], uint32(len(data)))
	return writer.WriteHeader(header[:])
}

func (c *grpcCodec) Decode(ctx context.Context, message remote.Message, in remote.ByteBuffer) (err error) {
	hdr, err := in.Next(5)
	if err != nil {
		return err
	}
	dLen := int(binary.BigEndian.Uint32(hdr[1:]))
	d, err := in.Next(dLen)
	if err != nil {
		return err
	}
	data := message.Data()
	switch t := data.(type) {
	case bprotoc.FastRead:
		_, err = bprotoc.Binary.ReadMessage(d, bprotoc.SkipTypeCheck, t)
		return err
	case protobufV2MsgCodec:
		return t.XXX_Unmarshal(d)
	case proto.Message:
		return proto.Unmarshal(d, t)
	case protobufMsgCodec:
		return t.Unmarshal(d)
	}
	return nil
}

func (c *grpcCodec) Name() string {
	return "grpc"
}
