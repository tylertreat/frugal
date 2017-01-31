package com.workiva.frugal.transport;

/**
 * FSubscriberTransportFactory produces FSubscriberTransports and is typically
 * used by an FScopeProvider.
 */
public interface FSubscriberTransportFactory {

    /**
     * Get a new FSubscriberTransport instance.
     *
     * @return A new FSubscriberTransport instance.
     */
    FSubscriberTransport getTransport();
}

