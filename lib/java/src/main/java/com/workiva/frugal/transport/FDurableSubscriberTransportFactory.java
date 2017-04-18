package com.workiva.frugal.transport;

/**
 * FDurableSubscriberTransportFactory produces FDurableSubscriberTransports and is typically
 * used by an FDurableScopeProvider.
 */
public interface FDurableSubscriberTransportFactory {

    /**
     * Get a new FDurableSubscriberTransport instance.
     *
     * @return A new FFDurableSubscriberTransport instance.
     */
    FDurableSubscriberTransport getTransport();
}

