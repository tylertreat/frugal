package com.workiva.frugal.protocol;

import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransportException;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;

import java.io.UnsupportedEncodingException;
import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.CyclicBarrier;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.ThreadPoolExecutor;
import java.util.concurrent.atomic.AtomicLong;
import java.util.stream.IntStream;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.fail;
import static org.mockito.Matchers.any;
import static org.mockito.Mockito.doAnswer;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.spy;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;

/**
 * Tests for {@link FRegistryImpl}.
 */
@RunWith(JUnit4.class)
public class FRegistryImplTest {

    /**
     * Returns a mock message frame.
     */
    private byte[] mockFrame(FContext context) throws TException, UnsupportedEncodingException {
        byte[] headers = HeaderUtils.encode(context.getRequestHeaders());
        byte[] message = "hello world".getBytes("UTF-8");
        byte[] frame = new byte[headers.length + message.length];
        System.arraycopy(headers, 0, frame, 0, headers.length);
        System.arraycopy(message, 0, frame, headers.length, message.length);
        return frame;
    }

    @Test
    public void testRegisterAddsToCallbackList() throws Exception {
        // given
        FRegistryImpl registry = new FRegistryImpl();
        FContext context = new FContext();

        // when
        registry.register(context, new ArrayBlockingQueue<>(1));

        // then
        assertEquals(1, registry.queueMap.size());
    }

    @Test(expected = TTransportException.class)
    public void testRegisterThrowsExceptionForMultipleAssignmentsToTheSameOpId() throws Exception {
        // given
        FRegistryImpl registry = new FRegistryImpl();
        FContext context = new FContext();

        // when
        registry.register(context, mock(BlockingQueue.class));

        // then (exception)
        registry.register(context, mock(BlockingQueue.class));
    }

    @Test
    public void testUnregisterRemovesFromHandlers() throws Exception {
        // given
        FRegistryImpl registry = new FRegistryImpl();
        FContext context = new FContext();
        BlockingQueue<byte[]> queue = new ArrayBlockingQueue<>(1);
        registry.register(context, queue);

        // when
        registry.unregister(context);

        // then
        assertEquals(0, registry.queueMap.size());
    }

    @Test
    public void testUnregisterMissingContextSucceeds() throws Exception {
        // given
        FRegistryImpl registry = new FRegistryImpl();

        // when
        registry.unregister(new FContext());
        registry.unregister(null);

        // then (succeeds)
    }

    @Test
    public void testExecuteDropsUnregisteredOpId() throws TException, UnsupportedEncodingException {
        // given
        FRegistryImpl registry = new FRegistryImpl();
        registry.queueMap = spy(registry.queueMap);

        FContext context = new FContext();
        byte[] frame = mockFrame(context);

        // when
        registry.execute(frame);

        // then
        verify(registry.queueMap, times(1)).get(any());
    }

    @Test
    public void testClosePutsPoisonPillInRegisteredQueues() throws Exception {
        // given
        CountDownLatch readySignal = new CountDownLatch(1);
        CountDownLatch interruptSignal = new CountDownLatch(1);

        FRegistryImpl registry = new FRegistryImpl();
        FContext context = new FContext();
        ThreadPoolExecutor executorService = (ThreadPoolExecutor) Executors.newCachedThreadPool();
        BlockingQueue<byte[]> queue = new ArrayBlockingQueue<>(1);
        executorService.execute(() -> {
            try {
                registry.register(context, queue); // no-op callback
            } catch (TTransportException e) {
                fail();
            }
            readySignal.countDown();

            // Wait for the queue to return the poison pill
            try {
                byte[] poisonPill = queue.take();
                assertEquals(FRegistry.POISON_PILL, poisonPill);
            } catch (InterruptedException e) {
                fail();
            }

            interruptSignal.countDown();
        });

        readySignal.await(); // wait for thread ready

        // when
        registry.close();

        // then (success when thread interrupted)
        interruptSignal.await(); // wait for thread interrupt
        assertEquals(registry.queueMap.size(), 0);
    }

    /**
     * Run a producer with multiple consumers.
     * All data put in to the registry must correctly be consumed to pass the test.
     * <p>
     * Note:
     * This test may unfairly synchronize consumers by pulling work from the same queue.
     * However, a shared-queue is indicative of real-world use.
     */
    @Test
    public void testRegistryIsThreadsafe() throws TException {
        final long poisonPill = Long.MAX_VALUE;

        final ExecutorService pool = Executors.newCachedThreadPool();

        // At test completion, values registered, unregistered, and executed must match
        final AtomicLong registerSum = new AtomicLong(0);
        final AtomicLong unregisterSum = new AtomicLong(0);
        final AtomicLong executeSum = new AtomicLong(0);

        final int nRegistrations = 100_000; // Number of registrations to make to the registry
        final int nConsumers = 100; // Number of concurrent consumers
        final CyclicBarrier barrier = new CyclicBarrier(nConsumers + 1 + 1); // + 1 producer, + 1 for main thread;
        final BlockingQueue<Long> opIds = new ArrayBlockingQueue<>(nRegistrations); // Store all operations registered
        final FRegistryImpl registry = new FRegistryImpl();

        class Producer implements Runnable {
            @Override
            public void run() {
                try {
                    barrier.await();

                    IntStream
                            .range(0, nRegistrations)
                            .forEach(i -> {
                                try {
                                    putToRegistry();
                                } catch (InterruptedException e) {
                                    fail();
                                }
                            });

                    // Signal end of queue with poison pill
                    opIds.add(poisonPill);

                    barrier.await();
                } catch (Exception e) {
                    throw new RuntimeException(e);
                }
            }

            private void putToRegistry() throws InterruptedException {
                FContext context = new FContext();

                BlockingQueue<byte[]> mockQueue = mock(BlockingQueue.class);
                doAnswer((invocationOnMock) -> {
                    executeSum.getAndAdd(context.getOpId());
                    return null;
                }).when(mockQueue).put(any());
                try {
                    registry.register(context, mockQueue);
                } catch (Exception e) {
                    throw new RuntimeException(e);
                }

                opIds.add(context.getOpId());
                registerSum.getAndAdd(context.getOpId());
            }

        }

        class Consumer implements Runnable {

            @Override
            public void run() {
                try {
                    barrier.await();

                    while (true) {
                        long opId = opIds.take();
                        if (opId == poisonPill) {
                            opIds.put(opId); // notify other threads to quit
                            barrier.await(); // release barrier
                            return;
                        }

                        FContext context = new FContext();
                        context.setOpId(opId);

                        byte[] frame = mockFrame(context);
                        registry.execute(frame);
                        registry.unregister(context);

                        unregisterSum.getAndAdd(context.getOpId());
                    }

                } catch (Exception e) {
                    throw new RuntimeException(e);
                }

            }

        }

        try {
            pool.execute(new Producer());
            IntStream.range(0, nConsumers).forEach(i -> pool.execute(new Consumer()));

            barrier.await(); // wait for all threads to be ready
            barrier.await(); // wait for all threads to finish

            assertEquals(registerSum.get(), unregisterSum.get());
            assertEquals(registerSum.get(), executeSum.get());
            assertEquals(registry.queueMap.size(), 0);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }

    }

}
