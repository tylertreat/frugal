package com.workiva.frugal;

import com.workiva.frugal.transport.FScopeTransport;
import com.workiva.frugal.transport.FScopeTransportFactory;

/**
 * FScopeProviders produce FScopeTransports and FProtocols for
 * use with Frugal Publishers and Subscribers.
 */
public class FScopeProvider {

    public class Client {
        private FScopeTransport transport;
        private FProtocol protocol;

        public Client(FScopeTransport t, FProtocol p) {
            transport = t;
            protocol = p;
        }

        public FScopeTransport getTransport() {
            return transport;
        }

        public FProtocol getProtocol() {
            return protocol;
        }
    }

    private FScopeTransportFactory transportFactory;
    private FProtocolFactory protocolFactory;

    public FScopeProvider(FScopeTransportFactory f, FProtocolFactory p) {
        transportFactory = f;
        protocolFactory = p;
    }

    /**
     * Returns a new Client containing a FScopeTransport and FProtocol used for pub/sub.
     *
     * @return Client with FScopeTransport and FProtocol.
     */
    public Client build() {
        FScopeTransport transport = transportFactory.getTransport();
        FProtocol protocol = protocolFactory.getProtocol(transport);
        return new Client(transport, protocol);
    }
}
