package com.workiva.frugal.exception;

import org.apache.thrift.TApplicationException;

/**
 * This exception indicates that a rate limit threshold has been exceeded.
 */
public class FRateLimitException extends TApplicationException {

    public static final int RATE_LIMIT_EXCEEDED = 102;

    public FRateLimitException() {
        super(RATE_LIMIT_EXCEEDED);
    }

    public FRateLimitException(String message) {
        super(RATE_LIMIT_EXCEEDED, message);
    }

    public FRateLimitException(int type, String message) {
        super(type, message);
    }

}
