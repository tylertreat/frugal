package com.workiva.frugal.middleware;

/**
 * ServiceMiddleware is used to apply middleware logic around service handlers.
 */
public interface ServiceMiddleware {

    /**
     * Handler contains a target handler, which may or may not be a dynamic proxy, and the service name.
     *
     * @param <T> the service handler type.
     */
    class Handler<T> {

        protected String service;
        protected T target;

        /**
         * Handler shouldn't be constructed by user code. User code should just pass through the Handler provided by
         * {@link ServiceMiddleware#apply(Handler)}.
         *
         * @param service the service name.
         * @param target  the service handler.
         */
        protected Handler(String service, T target) {
            this.service = service;
            this.target = target;
        }

    }

    /**
     * Returns an InvocationHandler which proxies the given Handler. This can be used to apply middleware logic around
     * a service call.
     *
     * @param next the next Handler in the chain.
     * @param <T>  the handler type.
     * @return proxied InvocationHandler.
     */
    <T> InvocationHandler<T> apply(Handler<T> next);

}
