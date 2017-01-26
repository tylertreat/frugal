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
}
