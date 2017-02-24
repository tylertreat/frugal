from tornado import gen


class FSubscription(object):
    """FSubscription to a pub/sub topic."""

    def __init__(self, topic, transport):
        """
        Initialize FSubscription. This is used only by generated code and should
        not be called directly.

        Args:
            topic: pub/sub topic string.
            transport: FSubscriberTransport
        """
        self._topic = topic
        self._transport = transport

    def get_topic(self):
        """Return subscription topic."""
        return self._topic

    @gen.coroutine
    def unsubscribe(self):
        """Unsubscribe from the topic."""
        yield self._transport.unsubscribe()
