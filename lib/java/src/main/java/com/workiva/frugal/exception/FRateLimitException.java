package com.workiva.frugal.exception;

import org.apache.thrift.TApplicationException;

/**
 * This exception indicates that a rate limit threshold has been exceeded.
 */
public class FRateLimitException extends TApplicationException {

    public FRateLimitException() {
        super(FApplicationException.RATE_LIMIT_EXCEEDED);
    }

    public FRateLimitException(String message) {
        super(FApplicationException.RATE_LIMIT_EXCEEDED, message);
    }

}
