package com.workiva.frugal.protocol;

import org.apache.thrift.protocol.TProtocolFactory;
import org.apache.thrift.transport.TTransport;

/**
 * FProtocolFactory creates FProtocols.
 */
public class FProtocolFactory {
    private TProtocolFactory tProtocolFactory;

    public FProtocolFactory(TProtocolFactory tProtocolFactory) {
        this.tProtocolFactory = tProtocolFactory;
    }

    public FProtocol getProtocol(TTransport transport) {
        return new FProtocol(tProtocolFactory.getProtocol(transport));
    }

}
