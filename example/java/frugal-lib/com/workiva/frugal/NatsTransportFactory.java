package com.workiva.frugal;

import org.nats.Connection;

/**
 * NatsTransportFactory is an implementation of the TransportFactory interface for
 * creating Transports backed by the NATS messaging system.
 */
public class NatsTransportFactory implements TransportFactory {

    private Connection conn;

    public NatsTransportFactory(Connection conn) {
        this.conn = conn;
    }

    @Override
    public Transport getTransport() {
        return new NatsTransport(conn);
    }

}
