library frugal.transport;

import "dart:async";
import "dart:typed_data";

import "package:thrift/thrift.dart";
import "package:messaging_sdk/messaging_sdk.dart";

part "nats_transport.dart";
part "nats_thrift_transport.dart";

/// Responsible for creating new Frugal Transports.
abstract class TransportFactory {
  Transport getTransport();
}

/// Wraps a Thrift TTransport which supports pub/sub.
abstract class Transport {
  /// Open the Transport to receive messages on the subscription.
  Future subscribe(String);

  /// Close the Transport to stop receiving messages on the subscription.
  Future unsubscribe();

  /// Prepare the Transport for publishing to the given topic.
  void preparePublish(String);

  /// Return the wrapped Thrift TTransport.
  TTransport thriftTransport();

  /// Wrap the underlying TTransport with the TTransport returned by the
  /// given TTransportFactory.
  Future applyProxy(TTransportFactory);
}