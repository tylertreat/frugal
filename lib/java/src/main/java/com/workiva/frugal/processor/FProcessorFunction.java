package com.workiva.frugal.processor;

import com.workiva.frugal.FContext;
import com.workiva.frugal.FProtocol;
import org.apache.thrift.TException;

/**
 * FProcessorFunction performs an operation on the provided input/output protocols.
 */
public interface FProcessorFunction {

    void process(FContext ctx, FProtocol in, FProtocol out) throws TException;

}
