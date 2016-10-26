import logging

from thrift.transport.TTransport import TMemoryBuffer
from tornado import gen
from tornado.locks import Lock

from frugal.context import _OP_ID
from frugal.exceptions import FException
from frugal.util.headers import _Headers

logger = logging.getLogger(__name__)


class FRegistry(object):
    """
    Registry is responsible for multiplexing received
    messages to the appropriate callback.
    """
    @gen.coroutine
    def register(self, context, callback):
        """Register a callback for a given FContext.

        Args:
            context: FContext to register.
            callback: function to register.
        """
        pass

    @gen.coroutine
    def unregister(self, context):
        """Unregister the callback for a given FContext.

        Args:
            context: FContext to unregister.
        """
        pass

    @gen.coroutine
    def execute(self, frame):
        """Dispatch a single Frugal message frame.

        Args:
            frame: an entire Frugal message frame.
        """
        pass


class FRegistryImpl(FRegistry):
    """
    FClientRegistry is intended for use only by Frugal clients.
    This is only to be used by generated code.
    """

    def __init__(self):
        self._handlers = {}
        self._handlers_lock = Lock()
        self._next_opid = 0
        self._opid_lock = Lock()

    @gen.coroutine
    def register(self, context, callback):
        """Register a callback for a given FContext.

        Args:
            context: FContext to register.
            callback: function to register.
        """
        # An FContext can be reused for multiple requests. Because of this,
        # every time an FContext is registered, it must be assigned a new op id
        # to ensure we can properly correlate responses. We use a monotonically
        # increasing atomic uint64 for this purpose. If the FContext already
        # has an op id, it has been used for a request. We check the handlers
        # map to ensure that request is not still in-flight.
        with (yield self._handlers_lock.acquire()):
            if str(context._get_op_id()) in self._handlers:
                raise FException("context already registered")

        op_id = yield self._increment_and_get_next_op_id()
        context._set_op_id(op_id)

        with (yield self._handlers_lock.acquire()):
            self._handlers[str(op_id)] = callback

    @gen.coroutine
    def unregister(self, context):
        """Unregister the callback for a given FContext.

        Args:
            context: FContext to unregister.
        """
        with (yield self._handlers_lock.acquire()):
            self._handlers.pop(str(context._get_op_id()), None)

    @gen.coroutine
    def execute(self, frame):
        """Dispatch a single Frugal message frame.

        Args:
            frame: an entire Frugal message frame.
        """
        if not frame:
            return
        headers = _Headers.decode_from_frame(frame)
        op_id = headers.get(_OP_ID, None)

        if not op_id:
            raise FException("Frame missing op_id")

        with (yield self._handlers_lock.acquire()):
            handler = self._handlers.get(op_id, None)
            if not handler:
                logger.warning("Got a message for unregistered context."
                               "Dropping")
                return

            handler(TMemoryBuffer(frame))

    @gen.coroutine
    def _increment_and_get_next_op_id(self):
        with (yield self._opid_lock.acquire()):
            self._next_opid += 1
            op_id = self._next_opid
        raise gen.Return(op_id)
