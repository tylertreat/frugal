import 'dart:async';
import 'dart:html';

import 'package:thrift/thrift.dart';
import 'package:event/event.dart' as event;
import 'package:frugal/frugal.dart' as frugal;
import 'package:messaging_frontend/messaging_frontend.dart' show Message, MessagingFrontendClient, Nats;
import 'package:w_transport/w_transport.dart';
import 'package:w_transport/w_transport_browser.dart' show configureWTransportForBrowser;

frugal.Subscription sub;

/// Adapted from the AS3 tutorial
void main() {
  new EventUI(querySelector('#output')).start();
}

class EventUI {
  final DivElement output;

  EventUI(this.output);

  event.EventsPublisher _eventsPublisher;
  event.EventsSubscriber _eventsSubscriber;

  frugal.FTransport _fTransport;
  event.FFooClient _fFooClient;

  void start() {
    _buildInterface();
    _initConnection();
  }

  void _initConnection() {
    configureWTransportForBrowser(useSockJS: true, sockJSProtocolsWhitelist: ["websocket", "xhr-streaming"]);
    var client = new MessagingFrontendClient("http://localhost:8100", "some-sweet-client", new Client());
    var nats = client.nats();
    nats.connect().then((_) {
      var provider = new frugal.Provider(new frugal.FNatsTransportFactory(nats), null, new TBinaryProtocolFactory());
      _eventsPublisher = new event.EventsPublisher(provider);
      _eventsSubscriber = new event.EventsSubscriber(provider);

      var timeout = new Duration(seconds: 1);
      frugal.TNatsTransportFactory.New(nats, "foo", timeout, timeout).then((TTransport T) {
        T.open().then((_){
          frugal.FProtocol protocol = new frugal.FBinaryProtocol(T);
          _fFooClient = new event.FFooClient(protocol);
        });
      });
    });
  }

  void _buildInterface() {
    output.children.forEach((e) {
      e.remove();
    });

    _buildPublishComponent();
    _buildSubscribeComponent();
    _buildRequestComponent();
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

  void _buildRequestComponent() {
    output.append(new HeadingElement.h3()
      ..text = "Foo Sevice");
    ButtonElement pingButton = new ButtonElement()
      ..text = "Ping"
      ..onClick.listen(_onPingClick);
    output.append(pingButton);
    InputElement blahMsg = new InputElement()
      ..id = "blahMsg"
      ..type = "number";
    output.append(blahMsg);
    ButtonElement blahButton = new ButtonElement()
      ..text = "Blah"
      ..onClick.listen(_onBlahClick);
    output.append(blahButton);
  }

  void _onPingClick(MouseEvent e) {
    var ctx = new frugal.Context("some-sweet-correlation");
    _fFooClient.ping(ctx);
  }

  void _onBlahClick(MouseEvent e) {
    var ctx = new frugal.Context("some-other-correlation");
    InputElement blahMsg = querySelector("#blahMsg");
    var num = int.parse(blahMsg.value);
    var e = new event.Event();
    e.message = "(╯°□°)╯︵ ┻━┻";
    _fFooClient.blah(ctx, num, "yey", e).then((int r) {
      window.alert("Got this rpc response ${r.toString()}");
    });
  }

  void onEvent(event.Event e) {
    window.alert(e.toString());
  }
}

