package server

import (
	"math/rand"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// あんま関係なさそう

const scopeName = "test/work"

var (
	meter                = otel.Meter(scopeName)
	tracer               = otel.Tracer(scopeName)
	sleepHistogram       metric.Float64Histogram
	subRequestsHistogram metric.Int64Histogram
)

func init() {
	var err error
	// [START opentelemetry_instrumentation_sleep_histogram_init]
	sleepHistogram, err = meter.Float64Histogram("example.sleep.duration",
		metric.WithDescription("Sample histogram to measure time spent in sleeping"),
		metric.WithExplicitBucketBoundaries(0.05, 0.075, 0.1, 0.125, 0.150, 0.2),
		metric.WithUnit("s"))
	if err != nil {
		panic(err)
	}
	// [END opentelemetry_instrumentation_sleep_histogram_init]

	subRequestsHistogram, err = meter.Int64Histogram("example.subrequests",
		metric.WithDescription("Sample histogram to measure the number of subrequests made"),
		metric.WithExplicitBucketBoundaries(1, 2, 3, 4, 5, 6, 7, 8, 9, 10),
		metric.WithUnit("{request}"))
	if err != nil {
		panic(err)
	}
}

func randomSleep(r *http.Request) time.Duration {
	// simulate the work by sleeping 100 to 200 ms
	sleepTime := time.Duration(100+rand.Intn(100)) * time.Millisecond
	time.Sleep(sleepTime)

	hostValue := attribute.String("host.value", r.Host)
	// custom histogram metric to record time slept in seconds
	sleepHistogram.Record(r.Context(), sleepTime.Seconds(), metric.WithAttributes(hostValue))
	return sleepTime
}

func computeSubrequests(r *http.Request, subRequests int) error {
	// Add custom span representing the work done for the subrequests
	ctx, span := tracer.Start(r.Context(), "subrequests")
	defer span.End()

	// Make specified number of http requests to the /single endpoint.
	for i := 0; i < subRequests; i++ {
		if err := callSingle(ctx); err != nil {
			return err
		}
	}
	// record number of sub-requests made
	subRequestsHistogram.Record(ctx, int64(subRequests))
	return nil
}
