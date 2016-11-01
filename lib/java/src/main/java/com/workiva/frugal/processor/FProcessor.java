package com.workiva.frugal.processor;

import com.workiva.frugal.middleware.ServiceMiddleware;
import com.workiva.frugal.protocol.FProtocol;
import org.apache.thrift.TException;

/**
 * FProcessor is Frugal's equivalent of Thrift's TProcessor. It's a generic object
 * which operates upon an input stream and writes to an output stream.
 * Specifically, an FProcessor is provided to an FServer in order to wire up a
 * service handler to process requests.
 */
public interface FProcessor {

    /**
     * Processes the request from the input protocol and write the response to the output protocol.
     *
     * @param in  input FProtocol
     * @param out output FProtocol
     * @throws TException
     */
    void process(FProtocol in, FProtocol out) throws TException;

    /**
     * Adds the given ServiceMiddleware to the FProcessor. This should only be called before the server is started.
     *
     * @param middleware the ServiceMiddleware to add
     */
    void addMiddleware(ServiceMiddleware middleware);
}
