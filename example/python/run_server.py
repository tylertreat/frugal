import logging
import sys
sys.path.append('gen-py.tornado')
sys.path.append('example_handler.py')

from thrift.protocol import TBinaryProtocol

from tornado import gen, ioloop

from nats.io.client import Client as NATS

from frugal.processor import FProcessorFactory
from frugal.protocol import FProtocolFactory
from frugal.tornado.server import FNatsTornadoServer
from frugal.tornado.transport import FMuxTornadoTransportFactory

from event.f_Foo import Processor as FFooProcessor
from example_handler import ExampleHandler


root = logging.getLogger()
root.setLevel(logging.DEBUG)

ch = logging.StreamHandler(sys.stdout)
ch.setLevel(logging.DEBUG)
formatter = logging.Formatter(
    '%(asctime)s - %(levelname)s - %(message)s')
ch.setFormatter(formatter)
root.addHandler(ch)


@gen.coroutine
def main():

    nats_client = NATS()
    options = {
        "verbose": True,
        "servers": ["nats://127.0.0.1:4222"]
    }

    yield nats_client.connect(**options)

    prot_factory = FProtocolFactory(TBinaryProtocol.TBinaryProtocolFactory())
    transport_factory = FMuxTornadoTransportFactory()

    handler = ExampleHandler()
    processor = FFooProcessor(handler)
    processor_factory = FProcessorFactory(processor)

    subject = "foo"
    heartbeat_interval = 10000
    max_missed_heartbeats = 3

    server = FNatsTornadoServer(nats_client,
                                subject,
                                max_missed_heartbeats,
                                processor_factory,
                                transport_factory,
                                prot_factory,
                                heartbeat_interval)

    logging.info("Starting server...")

    yield server.serve()

if __name__ == '__main__':
    io_loop = ioloop.IOLoop.instance()
    io_loop.add_callback(main)
    io_loop.start()
