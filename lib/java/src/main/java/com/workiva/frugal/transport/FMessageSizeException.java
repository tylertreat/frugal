package com.workiva.frugal.transport;

import org.apache.thrift.transport.TTransportException;


public class FMessageSizeException extends TTransportException {

    public FMessageSizeException() {
        super();
    }

    public FMessageSizeException(String message) {
        super(message);
    }

    public FMessageSizeException(Throwable cause) {
        super(cause);
    }

    public FMessageSizeException(String message, Throwable cause) {
        super(message, cause);
    }

}
