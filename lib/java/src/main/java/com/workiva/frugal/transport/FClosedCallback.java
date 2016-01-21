package com.workiva.frugal.transport;

/**
 * When a {@code FTransport} is closed for any reason, the {@code FTransport}
 * object's {@code FClosedCallback} is notifed, if one has been registered.
 */
public interface FClosedCallback {
    /**
     * This callback notification method is invoked when the {@code FTransport} is
     * closed.
     */
    void onClose();
}

