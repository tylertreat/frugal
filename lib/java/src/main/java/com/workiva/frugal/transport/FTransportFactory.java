package com.workiva.frugal.transport;

import org.apache.thrift.transport.TTransport;

/**
 * FTransportFactory is responsible for creating new FTransports.
 */
public interface FTransportFactory {

    /**
     * Returns a new FTransport wrapping the given TTransport.
     *
     * @param transport TTransport to wrap
     * @return new FTransport
     */
    FTransport getTransport(TTransport transport);
}
