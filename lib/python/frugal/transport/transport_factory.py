class FTransportFactory(object):
    """FTransportFactory is responsible for creating new FTransports."""

    def get_transport(self, thrift_transport):
        """ Retuns a new FTransport wrapping the given TTransport.

        Args:
            thrift_transport: TTransport to wrap.
        Returns:
            new FTranpsort
        """
        pass


class FPublisherTransportFactory:
    """
    FPublisherTransportFactory is responsible for creating new
    FPublisherTransports.
    """
    def get_transport(self):
        """Returns a new FPublisherTransport."""
        pass


class FSubscriberTransportFactory:
    """
    FSubscriberTransportFactory is responsible for creating new
    FSubscriberTransports.
    """
    def get_transport(self):
        """Returns a new FSubscriberTransport."""
        pass
