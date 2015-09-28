package com.workiva.frugal;

import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;

/**
 * NatsThriftTransport is an extension of thrift.TTransport exclusively used for
 * pub/sub via NATS.
 */
public class NatsThriftTransport extends TTransport {

    @Override
    public boolean isOpen() {
        // TODO
        return false;
    }

    @Override
    public void open() throws TTransportException {
        // TODO
    }

    @Override
    public void close() {
        // TODO
    }

    @Override
    public int read(byte[] bytes, int i, int i1) throws TTransportException {
        // TODO
        return 0;
    }

    @Override
    public void write(byte[] bytes, int i, int i1) throws TTransportException {
        // TODO
    }

    public void setSubject(String subject) {
        // TODO
    }

}
