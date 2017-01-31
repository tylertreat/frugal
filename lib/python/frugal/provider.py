
class FScopeProvider(object):
    """FScopeProviders produce FScopeTransports and FProtocols for use with
    Frugal Publishers and Subscribers. This also provides a shim for adding
    middleware to a publisher or subscriber.
    """

    def __init__(self, pub_transport_factory, sub_transport_factory,
                 protocol_factory, middleware=None):
        """Initialize FScopeProvider.

        Args:
            pub_transport_factory: FPublisherTransportFactory
            sub_transport_factory: FSubscriberTransportFactory
            protocol_factory: FProtocolFactory
            middleware: list of ServiceMiddleware
        """
        self._pub_transport_factory = pub_transport_factory
        self._sub_transport_factory = sub_transport_factory
        self._protocol_factory = protocol_factory
        self._middleware = middleware or []

    def new_publisher(self):
        """Return a tupled FScopeTransport and FProtocol.
        Returns:
            (FScopeTransport, FProtocolFactory)
        """
        transport = self._pub_transport_factory.get_transport()
        return transport, self._protocol_factory

    def new_subscriber(self):
        transport = self._sub_transport_factory.get_transport()
        return transport, self._protocol_factory

    def get_middleware(self):
        """Return the ServiceMiddleware stored on this FScopeProvider."""
        return list(self._middleware)


class FServiceProvider(object):
    """FServiceProvider is the service equivalent of FScopeProvider. It produces
     FTransports and FProtocols for use by RPC service clients. The main
     purpose of this is to provide a shim for adding middleware to a client.
     """

    def __init__(self, transport, protocol_factory, middleware=None):
        """Initialize FServiceProvider.

        Args:
            transport: FTransport
            protocol_factory: FProtocolFactory
            middleware: list of ServiceMiddleware
        """
        self._transport = transport
        self._protocol_factory = protocol_factory
        self._middleware = middleware or []

    def get_transport(self):
        """Get the FTransport from the provider."""
        return self._transport

    def get_protocol_factory(self):
        """Get the FProtocolFactory from the provider."""
        return self._protocol_factory

    def get_middleware(self):
        """Return the ServiceMiddleware stored on this FServiceProvider."""
        return list(self._middleware)
