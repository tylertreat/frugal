part of frugal;

/// Factory for Service TSocketTransport backed by NATS client
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
        throw new StateError("frugal: no reply subject on connect");
      }

      // Connect message consists of "[heartbeat subject] [heartbeat reply subject] [expected interval ms]"
      List<String> subjects = _codec.decode(msg.payload).split(" ");
      if (subjects.length != 3) {
        throw new StateError("frugal: invalid connect message");
      }

      String heartbeatListen = subjects[0];
      String heartbeatReply = subjects[1];
      int deadline = int.parse(subjects[2]);
      Duration interval = new Duration();
      if (deadline > 0) {
        interval = new Duration(milliseconds: deadline);
      }

      var socket = new TNatsSocket(client, inbox, msg.reply, heartbeat, readTimeout, interval);
      return new TClientSocketTransport(socket);
    });
  }
}
