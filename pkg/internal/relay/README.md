# relay

The relay package contains generated protobuf and gRPC code used by the forwardrelay and receivingrelay packages. These files are generated from proto/electrician_relay.proto.

## Responsibilities

- Provide gRPC service definitions for relay streaming.
- Define message formats and service interfaces.

## Notes

This package is generated and should not be edited manually. Modify the proto file and regenerate stubs instead.

## References

- proto definition: proto/electrician_relay.proto
- forward relay: pkg/internal/forwardrelay
- receiving relay: pkg/internal/receivingrelay
