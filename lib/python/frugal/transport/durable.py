class FDurablePublisherTransport(object):
    """FDurablePublisherTransport is used exclusively for pub/sub scopes.
    Publishers use it to publish messages durably to a topic.
    """

    def open(self):
        """Opens the transport."""
        pass

    def close(self):
        """Closes the transport."""
        pass

    def is_open(self):
        """Returns tru if the the transport is open, false otherwise.

        :return: true if open, false if closed
        :rtype: bool
        """
        pass

    def get_publish_size_limit(self):
        """Returns the maximum allowable size of a payload to be published.

        :return: publish size limit, negative number if unlimited
        :rtype: int
        """
        pass

    def publish(self, subject, group_id, payload):
        """Publish sens the given payload with the transport. Implementations of
        publish should be threadsafe.

        :param subject: subject to publish the message on
        :type subject: string
        :param group_id: arbitrary value to group messages together empty val
            indicates no grouping
        :type: string
        :param payload: the message payload to be sent
        :type payload: bytearray
        """
        pass


class FDurableSubscriberTransport(object):
    """FDurableSubscriberTransport is used exclusively for pub/sub scopes.
    Subscribers use it to durable subscribe to a pub/sub topic.
    """

    def subscribe(self, subject, callback):
        """Opens the transport and sets the subscribe topic.

        :param subject: the subject to subscribe to
        :type subject: string
        :param callback: function to call when a message is received
        :type callback: func
        """
        pass

    def unsubscribe(self):
        """Unsubscribes from the topic and closes the transport.
        """
        pass

    def is_subscribed(self):
        """Returns true if the transport is subscribed to a topic, false
        otherwise.

        :return: true if subscribed, false if not
        :rtype: bool
        """
        pass
