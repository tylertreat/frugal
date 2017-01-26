package com.workiva.frugal.exception;

/**
 * Contains TTransportException types used in frugal instantiated TTransportExceptions.
 */
public class FTransportExceptionType {

    /**
     * TTransportException code which indicates the request was too large for the transport.
     */
    public static final int REQUEST_TOO_LARGE = 100;

    /**
     * TTransportException code which indicates the response was too large for the transport.
     */
    public static final int RESPONSE_TOO_LARGE = 101;

}
