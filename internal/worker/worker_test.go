package worker

import (
	"context"
	"llm_gateway/internal/events"
	"llm_gateway/internal/job"
	"testing"

	"github.com/google/uuid"
)

type fakeSvc struct {
	claim        *job.Job
	claimErr     error
	succeededErr error

	succeeded *job.ResponsePayload
	failed    bool
	failCause error
}

func (f *fakeSvc) ClaimJobToRun(context.Context, uuid.UUID) (*job.Job, error) {
	return f.claim, f.claimErr
}

func (f *fakeSvc) MarkSucceeded(_ context.Context, _ uuid.UUID, r *job.ResponsePayload) error {
	f.succeeded = r
	return f.succeededErr
}

func (f *fakeSvc) MarkJobFailed(_ context.Context, j *job.Job, errMsg error) error {
	f.failed, f.failCause = true, errMsg
	return nil
}

type fakeRouter struct {
	res      *job.ResponsePayload
	err      error
	attempts int
}

func (f *fakeRouter) Complete(context.Context, job.RequestPayload) (*job.ResponsePayload, error) {
	f.attempts++
	return f.res, f.err
}

type fakePub struct {
	events []events.UsageEvent
}

func (f *fakePub) Publish(_ context.Context, e events.UsageEvent) error {
	f.events = append(f.events, e)
	return nil
}

func (f *fakePub) Close() error {
	return nil
}

func newTestWorker(svc *fakeSvc, rtr *fakeRouter, pub *fakePub) *Worker {
	return &Worker{
		svc:    svc,
		router: rtr,
		events: pub,
	}
}

func TestProcessJob_Success(t *testing.T) {
	j := &job.Job{ID: uuid.New(), CallerID: uuid.New(), Attempts: 1}
	svc := &fakeSvc{claim: j}
	rtr := &fakeRouter{res: &job.ResponsePayload{Text: "hi", Provider: "claude", InputTokens: 10, OutputTokens: 5}}
	pub := &fakePub{}

	w := newTestWorker(svc, rtr, pub)

	if err := w.processJob(context.Background(), j.ID); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if svc.succeeded == nil || svc.failed {
		t.Fatal("expected MarkSucceeded")
	}

	if len(pub.events) != 1 || !pub.events[0].Success || pub.events[0].Provider != "claude" {
		t.Fatalf("expected one success event, got %+v", pub.events)
	}

	if pub.events[0].CallerID != j.CallerID {
		t.Fatal("event must carry the job's caller")
	}
}

func TestProcessJob_AlreadyClaimedJob(t *testing.T) {
	svc := &fakeSvc{claim: nil}
	rtr := &fakeRouter{res: &job.ResponsePayload{Text: "hi", Provider: "claude", InputTokens: 10, OutputTokens: 5}}
	pub := &fakePub{}

	w := newTestWorker(svc, rtr, pub)

	if err := w.processJob(context.Background(), uuid.New()); err != nil {
		t.Fatalf("duplicate message should be acked, got %v", err)
	}

	if rtr.attempts != 0 {
		t.Fatal("router must not be called for an unclaimed job")
	}

	if len(pub.events) != 0 {
		t.Fatal("no event for a job we didn't run")
	}
}
