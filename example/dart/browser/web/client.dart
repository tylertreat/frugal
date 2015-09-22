import 'dart:html';

import 'package:http/browser_client.dart' as http;
import 'package:thrift/thrift.dart';
import 'package:event/event.dart' as event;
import 'package:frugal/frugal.dart';
import 'package:messaging_sdk/messaging_sdk.dart';

/// Adapted from the AS3 tutorial
void main() {
  new EventUI(querySelector('#output')).start();
}

class EventUI {
  final DivElement output;

  EventUI(this.output);

  Transport _transport;
  event.EventPublisher _eventPublisher;
  event.EventSubscriber _eventSubscriber;

  void start() {
    _buildInterface();
    _initConnection();
  }

  void _initConnection() {
    var client = new http.BrowserClient();
    var nats = new Nats("http://localhost:8100/nats", "fooclient", client);
    nats.connect().then((_) {
      _transport = new NatsTransport(nats);
      _eventPublisher = new event.EventPublisher(new NatsTransportFactory(nats),
      null, new TJsonProtocolFactory());
      _eventSubscriber = new event.EventSubscriber(new NatsTransportFactory(nats),
      null, new TJsonProtocolFactory());
    });
  }

  void _buildInterface() {
    output.children.forEach((e) {
      e.remove();
    });

    _buildPublishComponent();
    _buildSubscribeComponent();
  }

  void _buildPublishComponent() {
    output.append(new HeadingElement.h3()
      ..text = "Publish Event");
    ButtonElement publishButton = new ButtonElement()
      ..text = "Publish"
      ..onClick.listen(_onPublishClick);
    output.append(publishButton);
  }

  void _onPublishClick(MouseEvent e) {
    var e = new event.Event();
    e.message = "foo";
    _eventPublisher.publishEventCreated(e);
  }

  void _buildSubscribeComponent() {
    output.append(new HeadingElement.h3()
      ..text = "Subscribe Event");
    ButtonElement subscribeButton = new ButtonElement()
      ..text = "Subscribe"
      ..onClick.listen(_onSubscribeClick);
    output.append(subscribeButton);
  }

  void _onSubscribeClick(MouseEvent e) {
    _eventSubscriber.subscribeEventCreated(onEvent);
  }

  void onEvent(event.Event e) {
    print(e);
  }

}