from thrift.transport.TTransport import TTransportBase


class FTransport(object):
    """
    FTransport is comparable to Thrift's TTransport in that it represents the
    transport layer for frugal clients. However, frugal is callback based and
    sends only framed data. Due to this, instead of read, write, and flush
    methods, FTransport has a send method that sends framed frugal messages.
    To handle callback data, an FTransport also has an FRegistry, so it provides
    methods for registering and unregistering an FAsyncCallback to an FContext.
    """
    def open(self):
        """Open the transport."""
        raise NotImplementedError('You must override this')

    def close(self):
        """Close the transport."""
        raise NotImplementedError('You must override this')

    def is_open(self):
        """Return True if the transport is open, False otherwise."""
        raise NotImplementedError('You must override this')

    def register(self, context, callback):
        """
        Register a callback with a context.

        Args:
            context: The context to register.
            callback: The function associated with the given context.
        """
        raise NotImplementedError('You must override this')

    def unregister(self, context):
        """
        Unregister the given context.

        Args:
            context: The context to unregister.
        """
        raise NotImplementedError('You must override this')

    def set_monitor(self, monitor):
        """
        Set the transport monitor for the transport. This should only be used
        for "stateful" transports.

        Args:
            monitor: A transport monitor
        """
        raise NotImplementedError('You must override this')

    def send(self, data):
        """Transmits the given data."""
        raise NotImplementedError('You must override this')

    def get_request_size_limit(self):
        """
        Returns the maximum number of bytes that can be sent. A non-positive
        number is returned to indicate an unbounded allowable size.
        """
        raise NotImplementedError('You must override this')


class TSynchronousTransport(TTransportBase, object):
    """FSynchronousTransport is a Thrift TTransport for services which makes
    synchronous requests.
    """

    def set_timeout(self, timeout):
        """Set the request timeout.

        Args:
            timeout: request timeout in milliseconds.
        """
        pass

