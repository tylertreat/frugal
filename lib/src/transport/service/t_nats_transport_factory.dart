part of frugal;

/// Factory for Service TTransport backed by NATS client
class TNatsTransportFactory {
  static Utf8Codec _codec = new Utf8Codec();

  static Future<TSocketTransport> New(Nats client, String subject, Duration timeout,
                                readTimeout) async {
    var inbox = await client.newInbox();
    var stream = await client.subscribe(inbox);
    client.publish(subject, inbox, new Uint8List.fromList([]));
    return stream.first.timeout(timeout).catchError(
            (e) => client.unsubscribe(inbox)).then((Message msg) {
      client.unsubscribe(inbox);
      if (msg.reply == "") {
        throw new StateError("thrift_nats: no reply subject on connect");
      }

      var heatbeatAndDeadline = _codec.decode(msg.payload).split(" ");
      if (heatbeatAndDeadline.length != 2) {
        throw new StateError("thrift_nats: invalid connect message");
      }

      var heartbeat = heatbeatAndDeadline[0];
      var deadline = int.parse(heatbeatAndDeadline[1]);
      Duration interval = new Duration();
      if (deadline > 0) {
        deadline = (deadline - (deadline/4)).floor();
        interval = new Duration(milliseconds: deadline);
      }

      var socket = new TNatsSocket(client, inbox, msg.reply, heartbeat, readTimeout, interval);
      return new TClientSocketTransport(socket);
    });
  }
}
