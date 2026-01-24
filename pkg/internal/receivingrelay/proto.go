package receivingrelay

import (
	"fmt"
	"reflect"
	"strings"

	"google.golang.org/protobuf/proto"
)

func isProtoTarget[T any](out *T) bool {
	if _, ok := any(out).(proto.Message); ok {
		return true
	}
	if _, ok := any(*out).(proto.Message); ok {
		return true
	}
	return false
}

func decodeProtoInto[T any](payloadType string, b []byte, out *T) error {
	if m, ok := any(out).(proto.Message); ok {
		if err := validatePayloadType(payloadType, m); err != nil {
			return err
		}
		return proto.Unmarshal(b, m)
	}

	rv := reflect.ValueOf(out).Elem()
	if rv.IsValid() && rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		if m, ok := rv.Interface().(proto.Message); ok {
			if err := validatePayloadType(payloadType, m); err != nil {
				return err
			}
			return proto.Unmarshal(b, m)
		}
	}

	return fmt.Errorf("payload_encoding=PROTO requires T to be a protobuf message type; got %T", out)
}

func validatePayloadType(want string, msg proto.Message) error {
	if want == "" {
		return nil
	}

	want = normalizeTypeName(want)
	got := normalizeTypeName(string(msg.ProtoReflect().Descriptor().FullName()))

	if want != got {
		return fmt.Errorf("payload_type mismatch: want=%s got=%s", want, got)
	}
	return nil
}

func normalizeTypeName(s string) string {
	if i := strings.LastIndex(s, "/"); i >= 0 {
		s = s[i+1:]
	}
	return s
}
