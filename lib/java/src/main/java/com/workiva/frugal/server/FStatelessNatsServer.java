package com.workiva.frugal.server;

import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.FBoundedMemoryBuffer;
import com.workiva.frugal.transport.FTransport;
import com.workiva.frugal.transport.TNatsServiceTransport;
import com.workiva.frugal.util.BlockingRejectedExecutionHandler;
import com.workiva.frugal.util.ProtocolUtils;
import io.nats.client.Connection;
import io.nats.client.Message;
import io.nats.client.MessageHandler;
import io.nats.client.Subscription;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TMemoryInputTransport;
import org.apache.thrift.transport.TTransport;

import java.io.IOException;
import java.util.Arrays;
import java.util.concurrent.*;
import java.util.logging.Logger;

/**
 * An implementation of FServer which uses NATS as the underlying transport. Clients must connect with the
 * TStatelessNatsTransport.
 */
public class FStatelessNatsServer implements FServer {

    private static final int DEFAULT_WORK_QUEUE_LEN = 64;
    private static final Logger LOGGER = Logger.getLogger(FStatelessNatsServer.class.getName());

    private final Connection conn;
    private final FProcessor processor;
    private final FProtocolFactory inputProtoFactory;
    private final FProtocolFactory outputProtoFactory;
    private final String subject;
    private final String queue;
    private long highWatermark = FTransport.DEFAULT_WATERMARK;
    protected ExecutorService workerPool;
    private final BlockingQueue<Object> shutdown = new ArrayBlockingQueue<>(1);

    /**
     * Creates a new FStatelessNatsServer which receives requests on the given subject and queue.
     * The worker count controls the size of the thread pool used to process requests. This uses a provided queue
     * length. If the queue fills up, newly received requests will block to be placed on the queue. Configurable
     * load-shedding logic will be triggered if requests wait for too long, based on the high watermark. Clients must
     * connect with the TStatelessNatsTransport.
     *
     * @param conn         NATS connection
     * @param processor    FProcessor used to process requests
     * @param protoFactory FProtocolFactory used for input and output protocols
     * @param subject      NATS subject to receive requests on
     * @param queue        NATS queue group to receive requests on
     * @param workerCount  thread pool size
     */
    private FStatelessNatsServer(Connection conn, FProcessor processor, FProtocolFactory protoFactory,
                                 String subject, String queue, int workerCount, int queueLength) {
        this.conn = conn;
        this.processor = processor;
        this.inputProtoFactory = protoFactory;
        this.outputProtoFactory = protoFactory;
        this.subject = subject;
        this.queue = queue;
        BlockingQueue<Runnable> workQueue = new ArrayBlockingQueue<>(queueLength);
        this.workerPool = new ThreadPoolExecutor(1, workerCount, 30, TimeUnit.SECONDS, workQueue,
                new BlockingRejectedExecutionHandler());
    }

    /**
     * Builder for configuring and constructing FStatelessNatsServer instances.
     */
    public static class Builder {

        private final Connection conn;
        private final FProcessor processor;
        private final FProtocolFactory protoFactory;
        private final String subject;
        private String queue = "";
        private int workerCount = 1;
        private int queueLength = DEFAULT_WORK_QUEUE_LEN;

        /**
         * Creates a new Builder which creates FStatelessNatsServers that subscribe to the given NATS subject.
         *
         * @param conn         NATS connection
         * @param processor    FProcessor used to process requests
         * @param protoFactory FProtocolFactory used for input and output protocols
         * @param subject      NATS subject to receive requests on
         */
        public Builder(Connection conn, FProcessor processor, FProtocolFactory protoFactory, String subject) {
            this.conn = conn;
            this.processor = processor;
            this.protoFactory = protoFactory;
            this.subject = subject;
        }

        /**
         * Adds a NATS queue group to receive requests on to the Builder.
         *
         * @param queue NATS queue group
         * @return Builder
         */
        public Builder withQueue(String queue) {
            this.queue = queue;
            return this;
        }

        /**
         * Adds a worker count which controls the size of the thread pool used to process requests (defaults to 1).
         *
         * @param workerCount thread pool size
         * @return Builder
         */
        public Builder withWorkerCount(int workerCount) {
            this.workerCount = workerCount;
            return this;
        }

        /**
         * Adds a queue length which controls the size of the work queue buffering requests (defaults to 64).
         *
         * @param queueLength work queue length
         * @return Builder
         */
        public Builder withQueueLength(int queueLength) {
            this.queueLength = queueLength;
            return this;
        }

        /**
         * Creates a new configured FStatelessNatsServer.
         *
         * @return FStatelessNatsServer
         */
        public FStatelessNatsServer build() {
            return new FStatelessNatsServer(conn, processor, protoFactory, subject, queue, workerCount, queueLength);
        }

    }

    @Override
    public void serve() throws TException {
        Subscription sub = conn.subscribe(subject, queue, newRequestHandler());
        LOGGER.info("Frugal server running...");
        try {
            shutdown.take();
        } catch (InterruptedException ignored) {
        }
        LOGGER.info("Frugal server stopping...");

        try {
            sub.unsubscribe();
        } catch (IOException ignored) {
        }
    }

    @Override
    public void stop() throws TException {
        // Attempt to perform an orderly shutdown of the worker pool by trying to complete any in-flight requests.
        workerPool.shutdown();
        try {
            if (!workerPool.awaitTermination(30, TimeUnit.SECONDS)) {
                workerPool.shutdownNow();
            }
        } catch (InterruptedException e) {
            workerPool.shutdownNow();
            Thread.currentThread().interrupt();
        }

        // Unblock serving thread.
        try {
            shutdown.put(new Object());
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    @Override
    public synchronized void setHighWatermark(long watermark) {
        highWatermark = watermark;
    }

    private synchronized long getHighWatermark() {
        return highWatermark;
    }

    /**
     * Creates a new MessageHandler which is invoked when a request is received.
     */
    protected MessageHandler newRequestHandler() {
        return new MessageHandler() {
            @Override
            public void onMessage(Message message) {
                String reply = message.getReplyTo();
                if (reply == null || reply.isEmpty()) {
                    LOGGER.warning("Discarding invalid NATS request (no reply)");
                    return;
                }

                workerPool.submit(new Request(message.getData(), System.currentTimeMillis(), message.getReplyTo(),
                        getHighWatermark(), inputProtoFactory, outputProtoFactory, processor, conn));
            }
        };
    }

    /**
     * Runnable which encapsulates a request received by the server.
     */
    protected static class Request implements Runnable {

        protected final byte[] frameBytes;
        protected final long timestamp;
        protected final String reply;
        protected final long highWatermark;
        protected final FProtocolFactory inputProtoFactory;
        protected final FProtocolFactory outputProtoFactory;
        protected final FProcessor processor;
        protected final Connection conn;

        public Request(byte[] frameBytes, long timestamp, String reply, long highWatermark, FProtocolFactory inputProtoFactory,
                       FProtocolFactory outputProtoFactory, FProcessor processor, Connection conn) {
            this.frameBytes = frameBytes;
            this.timestamp = timestamp;
            this.reply = reply;
            this.highWatermark = highWatermark;
            this.inputProtoFactory = inputProtoFactory;
            this.outputProtoFactory = outputProtoFactory;
            this.processor = processor;
            this.conn = conn;
        }

        @Override
        public void run() {
            long duration = System.currentTimeMillis() - timestamp;
            if (duration > highWatermark) {
                LOGGER.warning("frame spent " + duration + "ms in the transport buffer, your consumer might be backed up");
            }
            process();
        }

        private void process() {
            // Read and process frame (exclude first 4 bytes which represent frame size).
            byte[] frame = Arrays.copyOfRange(frameBytes, 4, frameBytes.length);
            TTransport input = new TMemoryInputTransport(frame);
            FBoundedMemoryBuffer output = new FBoundedMemoryBuffer(TNatsServiceTransport.NATS_MAX_MESSAGE_SIZE);
            try {
                processor.process(inputProtoFactory.getProtocol(input), outputProtoFactory.getProtocol(output));
            } catch (TException e) {
                LOGGER.warning("error processing frame: " + e.getMessage());
                return;
            }

            if (output.length() == 0) {
                return;
            }

            // Add frame size (4-byte int32).
            byte[] response = new byte[output.length() + 4];
            ProtocolUtils.writeInt(output.length(), response, 0);
            System.arraycopy(output.getArray(), 0, response, 4, output.length());

            // Send response.
            try {
                conn.publish(reply, response);
            } catch (IOException e) {
                LOGGER.warning("failed to send response: " + e.getMessage());
            }
        }

    }

}
