package com.workiva.frugal.exception;

import com.workiva.frugal.transport.FTransport;
import org.apache.thrift.transport.TTransportException;

/**
 * This exception indicates that a rate limit threshold has been exceeded.
 */
public class FRateLimitException extends TTransportException {

    public FRateLimitException() {
        super(FTransport.RATE_LIMIT_EXCEEDED);
    }

    public FRateLimitException(String message) {
        super(FTransport.RATE_LIMIT_EXCEEDED, message);
    }

    public FRateLimitException(int type, String message) {
        super(type, message);
    }

    public FRateLimitException(Throwable cause) {
        super(FTransport.RATE_LIMIT_EXCEEDED, cause);
    }

    public FRateLimitException(String message, Throwable cause) {
        super(FTransport.RATE_LIMIT_EXCEEDED, message, cause);
    }
}
