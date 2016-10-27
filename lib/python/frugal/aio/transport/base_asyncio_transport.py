from io import BytesIO

from frugal.transport import FTransport


class FTransportBase(FTransport):
    """
    FTransportBase serves as a base class for FTransports,
    implementing a potentially size limited buffered write method.
    """
    def __init__(self, max_message_size: int):
        """
        Args:
            max_message_size: The maximum amount of data allowed to be buffered,
                              0 indicates unlimited size.
        """
        super().__init__()
        self._max_message_size = max_message_size
        self._wbuf = BytesIO()

    def is_open(self) -> bool:
        raise NotImplementedError('You must override this')

    async def open(self):
        raise NotImplementedError('You must override this')

    async def close(self):
        raise NotImplementedError('You must override this')

    async def send(self, data):
        raise NotImplementedError('You must override this')

    def get_request_size_limit(self) -> int:
        return self._max_message_size

