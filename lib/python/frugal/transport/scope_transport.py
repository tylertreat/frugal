class FPublisherTransport(object):
    """
    FPublisherTransport is used exclusively for pub/sub scopes. Publishers use
    it to publish to a topic.
    """
    def __init__(self, max_message_size):
        self._max_message_size = max_message_size

    def open(self):
        """Opens the transport for use."""
        raise NotImplementedError('You must override this')

    def close(self):
        """Closes the transport."""
        raise NotImplementedError('You must override this')

    def is_open(self):
        """Returns True if the transport is open, False otherwise."""
        raise NotImplementedError('You must override this')

    def get_publish_size_limit(self):
        """
        Returns the maximum allowable size of a payload to be published. A
        non-positive number is returned to indicate an unbounded allowable size.
        """
        return self._max_message_size

    def publish(self, topic, data):
        """
        Publish sends the given data with the transport to the given topic.
        Implementations of publish should be threadsafe
        """
        raise NotImplementedError('You must override this')

    def _check_publish_size(self, data):
        """Returns True if the data is of permissible size, False otherwise."""
        return len(data) > self._max_message_size > 0


class FSubscriberTransport(object):
    """
    FSubscriberTransport is used exclusively for pub/sub scopes. Subscribers use
    it to subscribe to a pub/sub topic.
    """
    def subscribe(self, topic, callback):
        """
        Subscribes to a pub/sub topic and executes the callback with each
        received message.
        """
        raise NotImplementedError('You must override this')

    def unsubscribe(self):
        """Unsubscribes from the current topic."""
        raise NotImplementedError('You must override this')

    def suspend(self):
        return self.unsubscribe()

    def is_subscribed(self):
        """
        Returns True if the transport is subscribed to a topic, False otherwise.
        """
        raise NotImplementedError('You must override this')

