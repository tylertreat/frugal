import logging

from thrift.Thrift import TApplicationException, TMessageType, TType
from tornado import gen
from tornado.locks import Lock

logger = logging.getLogger(__name__)


class FProcessorFunction(object):

    @gen.coroutine
    def process(self, ctx, iprot, oprot):
        pass


class FProcessor(object):
    """FProcessor is a generic object which operates upon an input stream and
    writes to some output stream.
    """
    @gen.coroutine
    def process(self, iprot, oprot):
        pass


class FBaseProcessor(FProcessor):

    def __init__(self):
        """Create new instance of FBaseProcessor that will process requests."""
        self._processor_function_map = {}
        self._write_lock = Lock()

    def add_to_processor_map(self, key, proc):
        """Register the given FProcessorFunction.

        Args:
            key: processor function name
            proc: FProcessorFunction
        """
        self._processor_function_map[key] = proc

    def get_write_lock(self):
        """Return the write lock."""
        return self._write_lock

    @gen.coroutine
    def process(self, iprot, oprot):
        """Process an input protocol and output protocol

        Args:
            iprot: input FProtocol
            oport: ouput FProtocol

        Raises:
            TApplicationException: if the processor does not know how to handle
                                   this type of function.
        """
        context = iprot.read_request_headers()
        name, _, _ = iprot.readMessageBegin()

        processor_function = self._processor_function_map.get(name)

        # If the function was in our dict, call process on it.
        if processor_function:
            try:
                ret = yield processor_function.process(context, iprot, oprot)
            except Exception:
                logging.exception('frugal: user handler code raised unhandled '
                                  'exception on request with correlation id {}'
                                  .format(context.get_correlation_id()))
                raise gen.Return(None)
            raise gen.Return(ret)

        iprot.skip(TType.STRUCT)
        iprot.readMessageEnd()

        ex = TApplicationException(TApplicationException.UNKNOWN_METHOD,
                                   "Unknown function: {0}".format(name))

        with (yield self._write_lock.acquire()):
            oprot.write_response_headers(context)
            oprot.writeMessageBegin(name, TMessageType.EXCEPTION, 0)
            ex.write(oprot)
            oprot.writeMessageEnd()
            oprot.trans.flush()

        logger.exception(ex)
        raise ex
