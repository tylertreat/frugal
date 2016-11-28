package com.workiva.frugal.exception;

import org.apache.thrift.transport.TTransportException;

/**
 * TTransportException which indicates a message was too large for a transport to handle.
 */
public class FMessageSizeException extends TTransportException {

    private FMessageSizeException(int type, String message) {
        super(type, message);
    }

    /**
     * Creates a new FMessageSizeException for an oversized request.
     *
     * @param message exception message
     * @return FMessageSizeException
     */
    public static FMessageSizeException request(String message) {
        return new FMessageSizeException(FTransportException.REQUEST_TOO_LARGE, message);
    }

    /**
     * Creates a new FMessageSizeException for an oversized response.
     *
     * @param message exception message
     * @return FMessageSizeException
     */
    public static FMessageSizeException response(String message) {
        return new FMessageSizeException(FTransportException.RESPONSE_TOO_LARGE, message);
    }
}
