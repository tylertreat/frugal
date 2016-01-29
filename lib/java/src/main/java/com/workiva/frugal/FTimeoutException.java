package com.workiva.frugal;

public class FTimeoutException extends FException {

    public FTimeoutException() { super(); }

    public FTimeoutException(String message) { super(message); }

    public FTimeoutException(Throwable cause) { super(cause); }

    public FTimeoutException(String message, Throwable cause) { super(message, cause); }
}
