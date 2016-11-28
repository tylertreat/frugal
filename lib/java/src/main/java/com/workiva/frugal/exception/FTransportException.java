package com.workiva.frugal.exception;

import org.apache.thrift.transport.TTransportException;

/**
 * Contains TTransportException constants.
 */
public class FTransportException extends TTransportException {

    /**
     * TTransportException code which indicates the request was too large for the transport.
     */
    public static final int REQUEST_TOO_LARGE = 100;

    /**
     * TTransportException code which indicates the response was too large for the transport.
     */
    public static final int RESPONSE_TOO_LARGE = 101;

}
