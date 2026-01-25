# proto

This directory contains the protobuf definitions for Electrician relays. The generated Go code lives under pkg/internal/relay and is used by the forward and receiving relay packages.

## Contents

- electrician_relay.proto: relay service definitions and message types

## Generation

Regenerate stubs with protoc when modifying the proto definition. The generated files are committed to keep consumers and internal packages in sync.

## Usage

The forwardrelay and receivingrelay packages depend on these definitions to stream payloads across services.

## License

Apache 2.0. See ../LICENSE.
