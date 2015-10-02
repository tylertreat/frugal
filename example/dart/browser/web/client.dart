import 'dart:async';
import 'dart:html';

import 'package:http/browser_client.dart' as http;
import 'package:thrift/thrift.dart';
import 'package:event/event.dart' as event;
import 'package:frugal/frugal.dart' as frugal;
import 'package:messaging_sdk/messaging_sdk.dart';

frugal.Subscription sub;

/// Adapted from the AS3 tutorial
void main() {
  new EventUI(querySelector('#output')).start();
}

class EventUI {
  final DivElement output;

  EventUI(this.output);

  frugal.Transport _transport;
  event.EventsPublisher _eventsPublisher;
  event.EventsSubscriber _eventsSubscriber;

  void start() {
    _buildInterface();
    _initConnection();
  }

  void _initConnection() {
    var client = new http.BrowserClient();
    var sessionHandler = new SessionHandler(
        "http://localhost:8100", "fooclient", client);
    var nats = new Nats("http://localhost:8100", sessionHandler);
    nats.connect().then((_) {
      _transport = new frugal.NatsTransport(nats);
      var provider = new frugal.Provider(new frugal.NatsTransportFactory(nats), null, new TJsonProtocolFactory());
      _eventsPublisher = new event.EventsPublisher(provider);
      _eventsSubscriber = new event.EventsSubscriber(provider);
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
    InputElement pubId = new InputElement()
      ..id = "pubId"
      ..type = "number";
    output.append(pubId);
    InputElement pubMsg = new InputElement()
      ..id = "pubMsg"
      ..type = "string";
    output.append(pubMsg);
    ButtonElement publishButton = new ButtonElement()
      ..text = "Publish"
      ..onClick.listen(_onPublishClick);
    output.append(publishButton);
  }

  void _onPublishClick(MouseEvent e) {
    InputElement pubId = querySelector("#pubId");
    InputElement pubMsg = querySelector("#pubMsg");
    var e = new event.Event();
    e.iD = int.parse(pubId.value);
    e.message = pubMsg.value;
    _eventsPublisher.publishEventCreated("barUser", e);
  }

  void _buildSubscribeComponent() {
    output.append(new HeadingElement.h3()
      ..text = "Subscribe Event");
    ButtonElement subscribeButton = new ButtonElement()
      ..text = "Subscribe"
      ..onClick.listen(_onSubscribeClick);
    output.append(subscribeButton);
    ButtonElement unsubscribeButton = new ButtonElement()
      ..text = "Unsubscribe"
      ..onClick.listen(_onUnsubscribeClick);
    output.append(unsubscribeButton);
  }

  Future _onSubscribeClick(MouseEvent e) async {
    if (sub == null ){
      sub = await _eventsSubscriber.subscribeEventCreated("barUser", onEvent);
    }
  }

  Future _onUnsubscribeClick(MouseEvent e) async {
    if (sub != null ){
      await sub.unsubscribe();
      sub = null;
    }
  }

  void onEvent(event.Event e) {
    window.alert(e.toString());
  }
}
