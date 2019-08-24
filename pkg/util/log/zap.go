package logutil

import (
	"bytes"
	"fmt"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Proto(name string, msg proto.Message) zapcore.Field {
	return zap.Object(name, zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
		return enc.AddReflected(name, &jsonpbLogMarshaler{pb: msg})
	}))
}

type jsonpbLogMarshaler struct {
	m  jsonpb.Marshaler
	pb proto.Message
}

func (m *jsonpbLogMarshaler) MarshalJSON() ([]byte, error) {
	b := &bytes.Buffer{}
	if err := m.m.Marshal(b, m.pb); err != nil {
		return nil, fmt.Errorf("failed to marshal jsonpb: %v", err)
	}
	return b.Bytes(), nil
}
