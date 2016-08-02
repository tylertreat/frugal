from frugal.aio.transport import FTransportBase
from frugal.context import FContext
from frugal.registry import FRegistry


class FRegistryTransport(FTransportBase):
    """
    FRegistryTransport implements registry manipulation methods.
    """
    def __init__(self, max_message_size):
        super().__init__(max_message_size)
        self._registry = None

    def set_registry(self, registry: FRegistry):
        """
        Set the registry for the transport.

        Args:
             registry: The registry to set, must be non-null.
        """
        if not registry:
            raise ValueError('registry must be non-null')

        if self._registry:
            return

        self._registry = registry

    def register(self, context: FContext, callback):
        """
        Register a callback with a context.

        Args:
            context: The context to register.
            callback: The function associated with the given context.
        """
        if not self._registry:
            raise ValueError('registry must be set')

        self._registry.register(context, callback)

    def unregister(self, context: FContext):
        """
        Unregister the given context.

        Args:
            context: The context to unregister.
        """
        if not self._registry:
            raise ValueError('registry must be set')

        self._registry.unregister(context)

    def execute_frame(self, frame):
        """
        Executes the callback associated with the data frame.

        Args:
            frame: The frame to be executed.
        """
        if not self._registry:
            raise ValueError('registry must be set')
        self._registry.execute(frame)
