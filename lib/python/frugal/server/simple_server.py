from threading import Lock

from thrift.transport.TTransport import TTransportException
from thrift.transport.TTransport import TFramedTransport

from . import FServer


class FSimpleServer(FServer):
    def __init__(self, processor_factory, server_transport, protocol_factory):
        self._processor_factory = processor_factory
        self._server_transport = server_transport
        self._protocol_factory = protocol_factory
        self._stopped = False
        self._stopped_mu = Lock()

    def serve(self):
        self._server_transport.listen()
        while True:
            with self._stopped_mu:
                if self._stopped:
                    return

            client = self._server_transport.accept()
            framed = TFramedTransport(client)
            iprot = self._protocol_factory.get_protocol(framed)
            oprot = self._protocol_factory.get_protocol(framed)
            processor = self._processor_factory.get_processor(framed)

            try:
                while True:
                    with self._stopped_mu:
                        if self._stopped:
                            break

                    processor.process(iprot, oprot)
            except TTransportException:
                continue
            except Exception as e:
                print(e)
                break

            framed.close()

    def stop(self):
        self._stopped = True
