package com.workiva.frugal.provider;

import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.FPublisherTransport;
import com.workiva.frugal.transport.FPublisherTransportFactory;
import com.workiva.frugal.transport.FSubscriberTransport;
import com.workiva.frugal.transport.FSubscriberTransportFactory;

/**
 * FScopeProvider produces FPublisherTransports, FSubscriberTransports, and
 * FProtocols for use by pub/sub scopes. It does this by wrapping an
 * FPublisherTransportFactory, an FSubscriberTransportFactory, and an
 * FProtocolFactory.
 */
public class FScopeProvider {

    /**
     * Publisher of this scope.
     */
    public static class Publisher {
        private FPublisherTransport transport;
        private FProtocolFactory protocolFactory;

        private Publisher(FPublisherTransport t, FProtocolFactory pf) {
            transport = t;
            protocolFactory = pf;
        }

        public FPublisherTransport getTransport() {
            return transport;
        }

        public FProtocolFactory getProtocolFactory() {
            return protocolFactory;
        }
    }

    /**
     * Subscriber of this scope.
     */
    public static class Subscriber {
        private FSubscriberTransport transport;
        private FProtocolFactory protocolFactory;

        public Subscriber (FSubscriberTransport t, FProtocolFactory pf) {
            transport = t;
            protocolFactory = pf;
        }

        public FSubscriberTransport getTransport() {
            return transport;
        }

        public FProtocolFactory getProtocolFactory() {
            return protocolFactory;
        }
    }

    private FPublisherTransportFactory publisherTransportFactory;
    private FSubscriberTransportFactory subscriberTransportFactory;
    private FProtocolFactory protocolFactory;

    public FScopeProvider(FPublisherTransportFactory ptf, FSubscriberTransportFactory stf,
                          FProtocolFactory pf) {
        publisherTransportFactory = ptf;
        subscriberTransportFactory = stf;
        protocolFactory = pf;
    }

    /**
     * Returns a new Publisher containing an FPublisherTransport and FProtocolFactory
     * used for publishing.
     *
     * @return Publisher with FPublisherTransport and FProtocol.
     */
    public Publisher buildPublisher() {
        FPublisherTransport transport = publisherTransportFactory.getTransport();
        return new Publisher(transport, protocolFactory);
    }

    /**
     * Returns a new Subscriber containing an FSubscriberTransport and FProtocolFactory
     * used for subscribing.
     *
     * @return SubscriberClient with FSubscriberTransport and FProtocol.
     */
    public Subscriber buildSubscriber() {
        FSubscriberTransport transport = subscriberTransportFactory.getTransport();
        return new Subscriber(transport, protocolFactory);
    }
}
