package grpc

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
)

// JSONCodec implements grpc.encoding.Codec using standard library JSON.
// This is needed because we cannot use protoc to generate protobuf binary
// codec code; all message types are plain Go structs with json tags.
//
// Both the gRPC server and clients must register this codec and set
// grpc.ForceCodec(jsonCodec{}) in DialOptions / ServerOption.
const JSONCodecName = "json"

type jsonCodec struct{}

func (jsonCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (jsonCodec) Name() string {
	return JSONCodecName
}

// init registers the JSON codec with gRPC's encoding registry so that it
// can be selected by content-type or by ForceCodec.
func init() {
	encoding.RegisterCodec(jsonCodec{})
}
