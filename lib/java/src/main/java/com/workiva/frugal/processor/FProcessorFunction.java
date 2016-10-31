package com.workiva.frugal.processor;

import com.workiva.frugal.protocol.FContext;
import com.workiva.frugal.protocol.FProtocol;
import org.apache.thrift.TException;

/**
 * FProcessorFunction is used internally by generated code. An FProcessor
 * registers an FProcessorFunction for each service method. Like FProcessor, an
 * FProcessorFunction exposes a single process call, which is used to handle a
 * method invocation.
 */
public interface FProcessorFunction {

    /**
     * Processes the request from the input protocol and write the response to the output protocol.
     *
     * @param ctx FContext
     * @param in  input FProtocol
     * @param out output FProtocol
     * @throws TException
     */
    void process(FContext ctx, FProtocol in, FProtocol out) throws TException;

}
