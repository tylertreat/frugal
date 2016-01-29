package com.workiva.frugal.processor;

import com.workiva.frugal.FProtocol;
import org.apache.thrift.TException;

/**
 * FProcessor is a generic object which operates upon an input stream and writes to some output stream.
 */
public interface FProcessor  {
    void process(FProtocol in, FProtocol out) throws TException;
}
