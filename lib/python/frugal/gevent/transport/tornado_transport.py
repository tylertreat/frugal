from tornado import gen, locks

from frugal.tornado.registry import FRegistryImpl
from frugal.transport import FTransport


class FTornadoTransport(FTransport):
    """ FTornadoTransport implements the buffered write data and registry
    interactions shared by all FTransports.
    """

    def __init__(self, max_message_size=1024 * 1024):
        super(FTornadoTransport, self).__init__()
        self._max_message_size = max_message_size

        self._registry = FRegistryImpl()

    @gen.coroutine
    def register(self, context, callback):
        """Register a provided FContext and callback function with the
        transport's internal FRegistry.

        Args:
            context: FContext to register.
            callback: function to register.

        Raises:
            StandardError: if registry has not been set.
        """
        yield self._registry.register(context, callback)

    @gen.coroutine
    def unregister(self, context):
        """Unregsiter the given context from the transports internal registry.

        Args:
            context: FContext to remove from the registry.

        Raises:
            StandardError: if registry has not been set.
        """
        yield self._registry.unregister(context)

    def is_open(self):
        raise NotImplementedError("You must override this.")

    @gen.coroutine
    def open(self):
        raise NotImplementedError("You must override this.")

    @gen.coroutine
    def close(self):
        raise NotImplementedError("You must override this.")

    def get_request_size_limit(self):
        return self._max_message_size

    @gen.coroutine
    def send(self, data):
        raise NotImplementedError('You must override this.')

    @gen.coroutine
    def execute_frame(self, frame):
        """Execute a frugal frame.
        NOTE: this frame must include the frame size.
        """
        yield self._registry.execute(frame[4:])

