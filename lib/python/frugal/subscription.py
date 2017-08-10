class FSubscription(object):
    """FSubscription to a pub/sub topic.  This is used only by generated code
    and should not be called directly."""

    def __init__(self, topic, transport):
        """Initialize FSubscription.

        Args:
            topic: pub/sub topic string.
            transport: FScopeTransport for the subscription.
        """
        self._topic = topic
        self._transport = transport

    def get_topic(self):
        """Return subscription topic."""
        return self._topic

    def unsubscribe(self):
        """
        Unsubscribe from the topic.

        The result of this is a future that should be awaited/yielded
        appropriately.
        """
        return self._transport.unsubscribe()

    def remove(self):
        """
        Unsubscribe and removes durably stored information on the broker,
        if applicable.
        """
        return self._transport.remove()
