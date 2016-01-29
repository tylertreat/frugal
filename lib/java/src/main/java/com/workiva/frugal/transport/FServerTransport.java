package com.workiva.frugal.transport;

import org.apache.thrift.TException;

/**
 * Provides client FTransports.
 */
public interface FServerTransport {

    void listen() throws TException;
    FTransport accept() throws TException;
    void close() throws TException;

    /**
     * Optional method implementation. This signals to the server transport
     * that it should break out of any accept() or listen() that it is currently
     * blocked on. This method, if implemented, MUST be thread safe, as it may
     * be called from a different thread context than the other FServerTransport
     * methods.
     */
    void interrupt() throws TException;

}
