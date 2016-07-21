package com.workiva.frugal.util;

import java.util.concurrent.RejectedExecutionException;
import java.util.concurrent.RejectedExecutionHandler;
import java.util.concurrent.ThreadPoolExecutor;

/**
 * Thread-pool rejection policy that blocks incoming requests.
 */
public class BlockingRejectedExecutionHandler implements RejectedExecutionHandler {

    @Override
    public final void rejectedExecution(final Runnable r, final ThreadPoolExecutor executor) {
        try {
            executor.getQueue().put(r);
        } catch (InterruptedException e) {
            throw new RejectedExecutionException("Interrupted while waiting to put the element", e);
        }
    }

}
