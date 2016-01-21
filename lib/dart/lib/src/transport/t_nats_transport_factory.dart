part of frugal;

/// Factory for Service TSocketTransport backed by NATS client
class TNatsTransportFactory {
  static Utf8Codec _codec = new Utf8Codec();

  static Future<TSocketTransport> New(Nats client, String subject, Duration timeout,
                                readTimeout) async {
    var inbox = await client.newInbox();
    var stream = await client.subscribe(inbox);
    client.publish(subject, inbox, new Uint8List(0));
    return stream.first.timeout(timeout).whenComplete(
            () => client.unsubscribe(inbox)).then((Message msg) {
      if (msg.reply == '') {
        throw new FError.withMessage('no reply subject on connect');
      }

      // Connect message consists of "[heartbeat subject] [heartbeat reply subject] [expected interval ms]"
      List<String> subjects = _codec.decode(msg.payload).split(" ");
      if (subjects.length != 3) {
        throw new FError.withMessage('invalid connect message');
      }

      String heartbeatListen = subjects[0];
      String heartbeatReply = subjects[1];
      var deadline;
      try {
        deadline = int.parse(subjects[2]);
      } catch (e) {
        throw new FError.withMessage('invalid heartbeat interval ${e.toString()}');
      }

      Duration interval = new Duration();
      if (deadline > 0) {
        interval = new Duration(milliseconds: deadline);
      }

      var socket = new TNatsSocket(client, inbox, msg.reply, heartbeatListen, heartbeatReply, readTimeout, interval);
      return new TClientSocketTransport(socket);
    });
  }
}
