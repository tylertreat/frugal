package com.workiva.frugal.protocol;

import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransport;

public interface FAsyncCallback {
    void onMessage(TTransport transport) throws TException;
}

