from frugal.aio.transport import FTransportBase
from frugal.context import FContext
from frugal.aio.registry import FRegistryImpl


class FRegistryTransport(FTransportBase):
    """
    FRegistryTransport implements registry manipulation methods.
    """
    def __init__(self, max_message_size):
        super().__init__(max_message_size)
        self._registry = FRegistryImpl()

    async def register(self, context: FContext, callback):
        """
        Register a callback with a context.

        Args:
            context: The context to register.
            callback: The function associated with the given context.
        """
        await self._registry.register(context, callback)

    async def unregister(self, context: FContext):
        """
        Unregister the given context.

        Args:
            context: The context to unregister.
        """
        await self._registry.unregister(context)

    async def execute_frame(self, frame):
        """
        Executes the callback associated with the data frame.

        Args:
            frame: The frame to be executed.
        """
        await self._registry.execute(frame)
