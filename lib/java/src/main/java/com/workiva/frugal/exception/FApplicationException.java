package com.workiva.frugal.exception;

import org.apache.thrift.TApplicationException;

/**
 * Contains TApplicationException constants.
 */
public class FApplicationException extends TApplicationException {

    /**
     * TApplicationException code which indicates the response was too large for the transport.
     */
    public static final int RESPONSE_TOO_LARGE = 100;

    /**
     * TApplicationException code which indicates a rate limit was exceeded.
     */
    public static final int RATE_LIMIT_EXCEEDED = 102;

}
