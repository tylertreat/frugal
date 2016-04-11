package com.workiva.frugal.middleware;

import java.lang.reflect.Method;
import java.lang.reflect.Proxy;

/**
 * InvocationHandler processes a service method invocation on a proxy instance and returns the result.
 *
 * @param <T> the service handler type.
 */
public abstract class InvocationHandler<T> implements java.lang.reflect.InvocationHandler {

    private final ServiceMiddleware.Handler<T> handler;

    /**
     * Creates a new InvocationHandler wrapping the given Handler.
     *
     * @param next the wrapped service handler.
     */
    public InvocationHandler(ServiceMiddleware.Handler<T> next) {
        handler = next;
    }

    @Override
    public Object invoke(Object proxy, Method method, Object[] args) throws Throwable {
        return invoke(handler.service, method, handler.target, args);
    }

    /**
     * Called when the middleware is invoked by a service call.
     *
     * @param service  the name of the service being invoked.
     * @param method   the method being invoked.
     * @param receiver the method receiver being invoked.
     * @param args     the method arguments.
     * @return the method return value.
     * @throws Throwable thrown by the wrapped method.
     */
    public abstract Object invoke(String service, Method method, Object receiver, Object[] args) throws Throwable;

    /**
     * Applies ServiceMiddleware to the provided target object by constructing a dynamic proxy.
     *
     * @param service    the name of the service.
     * @param target     the service handler.
     * @param iface      the interface class the handler implements.
     * @param middleware the middleware to apply.
     * @param <T>        the handler type.
     * @return handler proxy.
     */
    @SuppressWarnings("unchecked")
    public static <T> T composeMiddleware(String service, T target, Class iface, ServiceMiddleware[] middleware) {
        InvocationHandler<T> handler = new InvocationHandler<T>(new ServiceMiddleware.Handler<>(service, target)) {
            @Override
            public Object invoke(String service, Method method, Object receiver, Object[] args) throws Throwable {
                return method.invoke(receiver, args);
            }
        };

        ClassLoader classLoader = target.getClass().getClassLoader();
        Class[] ifaces = new Class[]{iface};
        T proxy = (T) Proxy.newProxyInstance(classLoader, ifaces, handler);
        for (ServiceMiddleware m : middleware) {
            handler = m.apply(new ServiceMiddleware.Handler<>(service, proxy));
            proxy = (T) Proxy.newProxyInstance(classLoader, ifaces, handler);
        }
        return proxy;
    }

}
