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
     * PublisherClient of this scope.
     */
    public class PublisherClient {
        private FPublisherTransport transport;
        private FProtocolFactory protocolFactory;

        public PublisherClient(FPublisherTransport t, FProtocolFactory pf) {
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
     * SubscriberClient of this scope.
     */
    public class SubscriberClient {
        private FSubscriberTransport transport;
        private FProtocolFactory protocolFactory;

        public SubscriberClient(FSubscriberTransport t, FProtocolFactory pf) {
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
     * Returns a new PublisherClient containing an FPublisherTransport and FProtocol
     * used for publishing.
     *
     * @return PublisherClient with FPublisherTransport and FProtocol.
     */
    public PublisherClient buildPublisher() {
        FPublisherTransport transport = publisherTransportFactory.getTransport();
        return new PublisherClient(transport, protocolFactory);
    }

    /**
     * Returns a new SubscriberClient containing an FSubscriberTransport and
     * FProtocolFactory used for subscribing.
     *
     * @return SubscriberClient with FSubscriberTransport and FProtocol.
     */
    public SubscriberClient buildSubscriber() {
        FSubscriberTransport transport = subscriberTransportFactory.getTransport();
        return new SubscriberClient(transport, protocolFactory);
    }
}
