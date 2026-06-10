# Future Await Race

This deterministic application starts two cooperative numeric tasks per round
and joins them in a third task through the `await` protocol. A separate task
uses `future_yield()` while waiting for a cancellation request, so the program
also checks the Future cancellation path without relying on scheduler order.

It is intentionally separate from Future Pipeline: that application measures a
bounded producer/worker/collector topology, while this one makes the Future
`Awaitable` interface and repeated task joins part of the observable work.
