package com.workiva.frugal.provider;

import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.FTransport;

/**
 * FServiceProviders produce FTransports and FProtocolFactories for
 * use with Frugal service clients.
 */
public class FServiceProvider {

    private FTransport transport;
    private FProtocolFactory protocolFactory;

    public FServiceProvider(FTransport transport, FProtocolFactory protocolFactory) {
        this.transport = transport;
        this.protocolFactory = protocolFactory;
    }

    /**
     * Get the FTransport from the provider.
     *
     * @return FTransport instance stored on the provider.
     */
    public FTransport getTransport() {
        return transport;
    }

    /**
     * Get the FProtocolFactory from the provider.
     *
     * @return FProtocolFactory stored on the provider.
     */
    public FProtocolFactory getProtocolFactory() {
        return protocolFactory;
    }
}
