library event.src.frug_event;

import 'dart:async';

import 'package:thrift/thrift.dart' as thrift;
import 'package:frugal/frugal.dart' as frugal;
import 'event.dart' as event;

const String delimiter = ".";

class EventPublisher {
  frugal.Transport transport;
  thrift.TProtocol protocol;
  int seqId;

  EventPublisher(frugal.TransportFactory t, thrift.TTransportFactory f, thrift.TProtocolFactory p) {
    var provider = new frugal.Provider(t, f, p);
    var tp = provider.newTransportProtocol();
    transport = tp.transport;
    protocol = tp.protocol;
    seqId = 0;
  }

  Future publishEventCreated(event.Event req) async {
    var op = "EventCreated";
    var prefix = "";
    var subject = "${prefix}Events${delimiter}$op";
    transport.preparePublish(subject);
    var oprot = protocol;
    seqId++;
    var msg = new thrift.TMessage(op, thrift.TMessageType.CALL, seqId);
    oprot.writeMessageBegin(msg);
    req.write(oprot);
    oprot.writeMessageEnd();
    return oprot.transport.flush();
  }
}

class EventSubscriber {
  frugal.Provider provider;

  EventSubscriber(frugal.TransportFactory t, thrift.TTransportFactory f, thrift.TProtocolFactory p) {
    provider = new frugal.Provider(t, f, p);
  }

  Future<frugal.Subscription> subscribeEventCreated(dynamic onEvent(event.Event e)) async {
    var op = "EventCreated";
    var prefix = "";
    var subject = "${prefix}Events${delimiter}$op";
    var tp = provider.newTransportProtocol();
    await tp.transport.subscribe(subject);
    tp.transport.signalRead.listen((_) {
      var e = _recvEventCreated(op, tp.protocol);
      onEvent(e);
    });
    return new frugal.Subscription(subject, tp.transport);
  }

  event.Event _recvEventCreated(String op, thrift.TProtocol iprot) {
    var tMsg = iprot.readMessageBegin();
    if (tMsg.name != op) {
      thrift.TProtocolUtil.skip(iprot, thrift.TType.STRUCT);
      iprot.readMessageEnd();
      throw new thrift.TApplicationError(
          thrift.TApplicationErrorType.UNKNOWN_METHOD, tMsg.name);
    }
    var req = new event.Event();
    req.read(iprot);
    iprot.readMessageEnd();
    return req;
  }
}