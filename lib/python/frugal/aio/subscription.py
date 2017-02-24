from frugal.aio.transport.nats_scope_transport import FNatsSubscriberTransport


class FSubscription(object):
    """FSubscription to a pub/sub topic."""

    def __init__(self, topic: str, transport: FNatsSubscriberTransport):
        """
        Initialize FSubscription. This is used only by generated code and should
        not be called directly.

        Args:
            topic: pub/sub topic string.
            transport: FSubscriberTransport
        """
        self._topic = topic
        self._transport = transport

    def get_topic(self) -> str:
        """Return subscription topic."""
        return self._topic

    async def unsubscribe(self):
        """Unsubscribe from the topic."""
        await self._transport.unsubscribe()
