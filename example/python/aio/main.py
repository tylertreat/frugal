import logging
import sys
sys.path.append('gen-py.asyncio')

from thrift.protocol import TBinaryProtocol
from thrift.transport.TTransport import TTransportException

# from tornado import ioloop, gen
import asyncio

# from nats.io.client import Client as NATS
from nats.aio.client import Client as NatsClient

from frugal.context import FContext
from frugal.protocol import FProtocolFactory
from frugal.provider import FScopeProvider
from frugal.aio.transport import (
    FNatsTransport,
    FHttpTransport,
    FNatsScopeTransportFactory,
)


from event.f_Events_publisher import EventsPublisher
from event.f_Foo import Client as FFooClient
from event.ttypes import Event


root = logging.getLogger()
root.setLevel(logging.DEBUG)

ch = logging.StreamHandler(sys.stdout)
ch.setLevel(logging.DEBUG)
formatter = logging.Formatter(
    '%(asctime)s - %(levelname)s - %(message)s')
ch.setFormatter(formatter)
root.addHandler(ch)


async def main():
    logging.info("Starting...")

    nats_client = NatsClient()
    options = {
        "verbose": True,
        "servers": ["nats://127.0.0.1:4222"]
    }

    logging.debug("Connecting to NATS")
    await nats_client.connect(**options)

    prot_factory = FProtocolFactory(TBinaryProtocol.TBinaryProtocolFactory())

    if "-client" in sys.argv or len(sys.argv) == 1:
        root.debug("Running FFooClient with NATS")
        await run_client(nats_client, prot_factory)
    if "-publisher" in sys.argv or len(sys.argv) == 1:
        root.debug("Running EventsPublisher")
        await run_publisher(nats_client, prot_factory)
    if "-http" in sys.argv:
        root.debug("Running FFooClient with NATS and HTTP")
        await run_client(nats_client, prot_factory, http=True)

    await nats_client.close()


async def run_client(nats_client, prot_factory, http=False):
    nats_transport = FNatsTransport(nats_client, "foo")

    try:
        await nats_transport.open()
    except TTransportException as ex:
        root.error(ex)
        return

    foo_client = FFooClient(nats_transport, prot_factory,
                            middleware=logging_middleware)

    root.info('oneWay()')
    await foo_client.oneWay(FContext(), 99, {99: "request"})

    root.info('basePing()')
    await foo_client.basePing(FContext(timeout=5 * 1000))

    root.info('ping()')
    await foo_client.ping(FContext())

    ctx = FContext()
    event = Event(42, "hello world")
    root.info('blah()')
    b = await foo_client.blah(ctx, 100, "awesomesauce", event)
    root.info('Blah response {}'.format(b))
    root.info('Response header foo: {}'.format(ctx.get_response_header("foo")))

    await nats_transport.close()

    if not http:
        return

    http_transport = FHttpTransport('http://localhost:8090/frugal')

    try:
        await http_transport.open()
    except TTransportException as ex:
        logging.error(ex)
        return

    foo_client = FFooClient(http_transport, prot_factory,
                            middleware=logging_middleware)
    print('oneWay()')
    await foo_client.oneWay(FContext(), 123, {123: 'request'})

    print('basePing()')
    await foo_client.basePing(FContext())

    print('ping()')
    await foo_client.ping(FContext())

    ctx = FContext()
    event = Event(43, 'other hello world')
    print('blah()')
    b = await foo_client.blah(ctx, 203, 'an http message', event)
    print('blah response {}'.format(b))
    print('response header foo: {}'.format(ctx.get_response_header('foo')))

    await http_transport.close()


async def run_publisher(nats_client, prot_factory):
    scope_transport_factory = FNatsScopeTransportFactory(nats_client)
    provider = FScopeProvider(scope_transport_factory, prot_factory)

    publisher = EventsPublisher(provider, middleware=logging_middleware)
    await publisher.open()

    event = Event(42, "hello, world!!!")
    publisher.publish_EventCreated(FContext(), "barUser", event)
    await publisher.close()


def logging_middleware(next):
    def handler(method, args):
        # service = '%s.%s' % (method.im_self.__module__,
        #                      method.im_class.__name__)
        # print '==== CALLING %s.%s ====' % (service, method.im_func.func_name)
        print('==== CALLING %s ====', method.__name__)
        ret = next(method, args)
        # print '==== CALLED  %s.%s ====' % (service, method.im_func.func_name)
        print('==== CALLED  %s ====', method.__name__)
        return ret
    return handler


if __name__ == '__main__':
    # Since we can exit after the client calls use `run_sync`
    # ioloop.IOLoop.instance().run_sync(main)
    io_loop = asyncio.get_event_loop()
    io_loop.run_until_complete(main())
