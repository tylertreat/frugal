package com.workiva.frugal.exception;

import org.apache.thrift.TException;

/**
 * Basic Frugal exception.
 */
public class FException extends TException {

    /**
     * TApplicationException code which indicates the response was too large for the transport.
     */
    public static final int RESPONSE_TOO_LARGE = 100;

    public FException() {
        super();
    }

    public FException(String message) {
        super(message);
    }

    public FException(Throwable cause) {
        super(cause);
    }

    public FException(String message, Throwable cause) {
        super(message, cause);
    }
}
