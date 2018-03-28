package com.workiva;

import com.workiva.frugal.middleware.InvocationHandler;
import com.workiva.frugal.middleware.ServiceMiddleware;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.server.FServlet;
import frugal.test.FFrugalTest;
import org.apache.thrift.protocol.TProtocolFactory;
import org.eclipse.jetty.server.Server;
import org.eclipse.jetty.servlet.ServletContextHandler;
import org.eclipse.jetty.servlet.ServletHolder;

import javax.servlet.Servlet;

import java.lang.reflect.Method;
import java.util.Arrays;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;

import static com.workiva.Utils.whichProtocolFactory;

public class TestServletServer {
    public static void main(String[] args) {
        try {
            CrossTestsArgParser parser = new CrossTestsArgParser(args);
            int port = parser.getPort();
            String protocolType = parser.getProtocolType();

            TProtocolFactory protocolFactory = whichProtocolFactory(protocolType);
            FProtocolFactory fProtocolFactory = new FProtocolFactory(protocolFactory);

            // Hand the transport to the handler
            FFrugalTest.Iface handler = new com.workiva.FrugalTestHandler();
            CountDownLatch called = new CountDownLatch(1);
            FFrugalTest.Processor processor = new FFrugalTest.Processor(handler, new TestServletServer.ServerMiddleware(called));

            Servlet servlet = new FServlet(processor, fProtocolFactory);
            ServletThread thread = new ServletThread(port, servlet);
            thread.start();

            // Wait for the middleware to be invoked, fail if it exceeds the longest client timeout (currently 20 sec)
            if (called.await(20, TimeUnit.SECONDS)) {
                System.out.println("Server middleware called successfully");
            } else {
                System.out.println("Server middleware not called within 20 seconds");
                System.exit(1);
            }
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    public static class ServletThread extends Thread {
        final int port;
        final Servlet servlet;

        public ServletThread(int port, Servlet servlet) {
            this.port = port;
            this.servlet = servlet;
        }

        public void run() {
            Server server = new Server(port);
            ServletContextHandler servletContextHandler = new ServletContextHandler(0);
            servletContextHandler.addServlet(new ServletHolder(servlet), "/");
            server.setHandler(servletContextHandler);

            try {
                server.start();
            } catch (Exception e) {
                e.printStackTrace();
            }
        }
    }

    private static class ServerMiddleware implements ServiceMiddleware {
        CountDownLatch called;

        ServerMiddleware(CountDownLatch called) {
            this.called = called;
        }

        @Override
        public <T> InvocationHandler<T> apply(T next) {
            return new InvocationHandler<T>(next) {
                @Override
                public Object invoke(Method method, Object receiver, Object[] args) throws Throwable {
                    Object[] subArgs = Arrays.copyOfRange(args, 1, args.length);
                    System.out.printf("%s(%s)\n", method.getName(), Arrays.toString(subArgs));
                    if (method.getName().equals("testOneway")) {

                        called.countDown();
                    }
                    return method.invoke(receiver, args);
                }
            };
        }
    }
}
