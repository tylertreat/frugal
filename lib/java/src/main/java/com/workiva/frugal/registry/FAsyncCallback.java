package com.workiva.frugal.registry;

import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransport;

public interface FAsyncCallback {
    void onMessage(TTransport transport) throws TException;
}

