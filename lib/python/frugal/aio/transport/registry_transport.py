from frugal.aio.transport import FAsyncIOTransportBase
from frugal.registry import FRegistry


class FAsyncIORegistryTransport(FAsyncIOTransportBase):
    # TODO locks?
    def __init__(self, max_message_size):
        super().__init__(max_message_size)
        self._registry = None

    def set_registry(self, registry: FRegistry):
        if self._registry:
            return

        if not registry:
            raise ValueError('registry must be set')

        self._registry = registry

    def register(self, context, callback):
        if not self._registry:
            raise ValueError('registry must be set')

        self._registry.register(context, callback)

    def unregister(self, context):
        if not self._registry:
            raise ValueError('registry must be set')

        self._registry.unregister(context)

    def execute(self, frame):
        if not self._registry:
            raise ValueError('registry must be set')
        self._registry.execute(frame)
