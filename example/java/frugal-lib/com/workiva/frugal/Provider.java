package com.workiva.frugal;

import org.apache.thrift.protocol.TProtocol;
import org.apache.thrift.protocol.TProtocolFactory;
import org.apache.thrift.transport.TTransportFactory;

/**
 * Providers produces Frugal Transports and Thrift TProtocols.
 */
public class Provider {

    public class Client {
        Transport transport;
        TProtocol protocol;

        public Client(Transport t, TProtocol p) {
            transport = t;
            protocol = p;
        }

        public Transport getTransport() {
            return transport;
        }

        public TProtocol getProtocol() {
            return protocol;
        }
    }

    private TransportFactory transportFactory;
    private TTransportFactory thriftTransportFactory;
    private TProtocolFactory protocolFactory;

    public Provider(TransportFactory t, TTransportFactory f, TProtocolFactory p) {
        transportFactory = t;
        thriftTransportFactory = f;
        protocolFactory = p;
    }

    /**
     * Returns a new Client containing a Transport and TProtocol used for pub/sub.
     *
     * @return Client with Transport and TProtocol.
     */
    public Client build() {
        Transport transport = transportFactory.getTransport();
        if (thriftTransportFactory != null) {
            transport.applyProxy(thriftTransportFactory);
        }
        TProtocol protocol = protocolFactory.getProtocol(transport.thriftTransport());
        return new Client(transport, protocol);
    }

}
