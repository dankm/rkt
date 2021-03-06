# rkt api-service

## Overview

The API service is designed to help users to get a well defined and structual result for listing and introspecting their pods and images.
The API service is implemented with [gRPC](http://www.grpc.io/).
The API service is designed to run without root privileges, and currently provides a read-only interface.
The API service is optional for running pods, the start/stop/crash of the API service won't affect any pods or images.

## Running the API service

The API service listens for gRPC requests on the address and port specified by the `--listen` option.
The default is to listen on the loopback interface on port number `15441`, equivalent to invoking `rkt api-service --listen=localhost:15441`.
Specify the address `0.0.0.0` to listen on all interfaces.

## Using the API service

The interfaces are defined in the [protobuf here](../../api/v1alpha/api.proto).
Here is a small [Go program](../../api/v1alpha/client_example.go) that illustrates how to use the API service.
