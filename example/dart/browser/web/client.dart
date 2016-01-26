import 'dart:async';
import 'dart:html';

import 'package:logging/logging.dart';
import 'package:thrift/thrift.dart';
import 'package:event/event.dart' as event;
import 'package:frugal/frugal.dart' as frugal;

frugal.FSubscription sub;

void main() {
  Logger.root.level = Level.FINEST;
  Logger.root.onRecord.listen((LogRecord r) {
    window.console.log('${r.loggerName}(${r.level}): ${r.message}');
  });
  new EventUI(querySelector('#output')).start();
}

class EventUI {
  final DivElement output;

  EventUI(this.output);

  event.EventsPublisher _eventsPublisher;
  event.EventsSubscriber _eventsSubscriber;

  event.FFooClient _fFooClient;

  void start() {
    _buildInterface();
    _initConnection();
  }

  void _initConnection() { }

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
    frugal.FContext ctx = new frugal.FContext(correlationId: 'an-id');
    _eventsPublisher.publishEventCreated(ctx, "barUser", e);
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
    if (_fFooClient == null) {
      window.alert("Not connected to server");
    }
    var ctx = new frugal.FContext(correlationId:"some-sweet-correlation");
    _fFooClient.ping(ctx).catchError( (e) {
      window.alert("Ping errored! ${e.toString()}");
    });
  }

  void _onBlahClick(MouseEvent e) {
    if (_fFooClient == null) {
      window.alert("Not connected to server");
    }
    var ctx = new frugal.FContext(correlationId: "some-other-correlation");
    InputElement blahMsg = querySelector("#blahMsg");
    var num = int.parse(blahMsg.value);
    var e = new event.Event();
    e.message = "(╯°□°)╯︵ ┻━┻";
    _fFooClient.blah(ctx, num, "yey", e).then((int r) {
      window.alert("Got this rpc response ${r.toString()}");
    });
  }

  void onEvent(frugal.FContext ctx, event.Event e) {
    window.alert(ctx.opId().toString() + ' : ' + e.toString());
  }
}

