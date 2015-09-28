package com.workiva.frugal;

import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.apache.thrift.transport.TTransportFactory;
import org.nats.Connection;

/**
 * NatsTransport is an implementation of the Transport interface backed by the NATS
 * messaging system.
 */
public class NatsTransport implements Transport {

    private TTransport thriftTransport;
    private NatsThriftTransport nats;

    public NatsTransport(Connection conn) {
        NatsThriftTransport tr = new NatsThriftTransport(conn);
        thriftTransport = tr;
        nats = tr;
    }

    @Override
    public void subscribe(String topic) throws TTransportException {
        nats.setSubject(topic);
        thriftTransport.open();
    }

    @Override
    public void unsubscribe() throws TTransportException {
        thriftTransport.close();
    }

    @Override
    public void preparePublish(String topic) {
        nats.setSubject(topic);
    }

    @Override
    public TTransport thriftTransport() {
        return thriftTransport;
    }

    @Override
    public void applyProxy(TTransportFactory proxy) {
        thriftTransport = proxy.getTransport(thriftTransport);
    }

    @Override
    public void flush() throws TTransportException {
        thriftTransport.flush();
    }

}
