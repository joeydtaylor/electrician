package websocketclient

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"google.golang.org/protobuf/proto"
	"nhooyr.io/websocket"
)

func normalizeFormat(format string) string {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		return types.WebSocketFormatJSON
	}
	return format
}

func validateFormat(format string) error {
	switch normalizeFormat(format) {
	case types.WebSocketFormatJSON, types.WebSocketFormatBinary, types.WebSocketFormatProto, types.WebSocketFormatText:
		return nil
	default:
		return fmt.Errorf("unsupported websocket format: %q", format)
	}
}

func encodeMessage[T any](format string, msg T) (websocket.MessageType, []byte, error) {
	switch normalizeFormat(format) {
	case types.WebSocketFormatJSON:
		b, err := json.Marshal(msg)
		return websocket.MessageText, b, err
	case types.WebSocketFormatText:
		b, err := encodeText(msg)
		return websocket.MessageText, b, err
	case types.WebSocketFormatBinary:
		b, err := encodeBinary(msg)
		return websocket.MessageBinary, b, err
	case types.WebSocketFormatProto:
		b, err := encodeProto(msg)
		return websocket.MessageBinary, b, err
	default:
		return websocket.MessageText, nil, fmt.Errorf("unsupported websocket format: %q", format)
	}
}

func decodeMessage[T any](format string, payload []byte) (T, error) {
	switch normalizeFormat(format) {
	case types.WebSocketFormatJSON:
		var out T
		if err := json.Unmarshal(payload, &out); err != nil {
			return out, err
		}
		return out, nil
	case types.WebSocketFormatText:
		return decodeText[T](payload)
	case types.WebSocketFormatBinary:
		return decodeBinary[T](payload)
	case types.WebSocketFormatProto:
		return decodeProto[T](payload)
	default:
		var out T
		return out, fmt.Errorf("unsupported websocket format: %q", format)
	}
}

func encodeText[T any](msg T) ([]byte, error) {
	switch v := any(msg).(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		return nil, fmt.Errorf("text format requires T to be string or []byte (got %T)", msg)
	}
}

func decodeText[T any](payload []byte) (T, error) {
	var out T
	switch any(out).(type) {
	case string:
		return any(string(payload)).(T), nil
	case []byte:
		copyBytes := append([]byte(nil), payload...)
		return any(copyBytes).(T), nil
	default:
		return out, fmt.Errorf("text format requires T to be string or []byte (got %T)", out)
	}
}

func encodeBinary[T any](msg T) ([]byte, error) {
	switch v := any(msg).(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("binary format requires T to be []byte or string (got %T)", msg)
	}
}

func decodeBinary[T any](payload []byte) (T, error) {
	var out T
	switch any(out).(type) {
	case []byte:
		copyBytes := append([]byte(nil), payload...)
		return any(copyBytes).(T), nil
	case string:
		return any(string(payload)).(T), nil
	default:
		return out, fmt.Errorf("binary format requires T to be []byte or string (got %T)", out)
	}
}

func encodeProto[T any](msg T) ([]byte, error) {
	if m, ok := any(msg).(proto.Message); ok {
		return proto.Marshal(m)
	}
	if m, ok := any(&msg).(proto.Message); ok {
		return proto.Marshal(m)
	}
	return nil, fmt.Errorf("proto format requires T to be a protobuf message (got %T)", msg)
}

func decodeProto[T any](payload []byte) (T, error) {
	var out T
	if m, ok := any(out).(proto.Message); ok {
		if err := proto.Unmarshal(payload, m); err != nil {
			return out, err
		}
		return out, nil
	}

	rv := reflect.ValueOf(&out).Elem()
	if rv.IsValid() && rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		if m, ok := rv.Interface().(proto.Message); ok {
			if err := proto.Unmarshal(payload, m); err != nil {
				return out, err
			}
			return out, nil
		}
	}

	return out, fmt.Errorf("proto format requires T to be a protobuf message (got %T)", out)
}
