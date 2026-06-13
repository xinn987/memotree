package buckets

import (
	"context"
	"reflect"
	"testing"
)

func TestEnsureAllEnsuresConfiguredBuckets(t *testing.T) {
	ensurer := &fakeEnsurer{}

	err := EnsureAll(context.Background(), ensurer, []string{"memotree-originals", "memotree-previews"})

	if err != nil {
		t.Fatalf("ensure buckets: %v", err)
	}
	want := []string{"memotree-originals", "memotree-previews"}
	if !reflect.DeepEqual(ensurer.buckets, want) {
		t.Fatalf("expected buckets %v, got %v", want, ensurer.buckets)
	}
}

func TestEnsureAllSkipsEmptyAndDuplicateBuckets(t *testing.T) {
	ensurer := &fakeEnsurer{}

	err := EnsureAll(context.Background(), ensurer, []string{" memotree-originals ", "", "memotree-originals"})

	if err != nil {
		t.Fatalf("ensure buckets: %v", err)
	}
	want := []string{"memotree-originals"}
	if !reflect.DeepEqual(ensurer.buckets, want) {
		t.Fatalf("expected buckets %v, got %v", want, ensurer.buckets)
	}
}

type fakeEnsurer struct {
	buckets []string
}

func (f *fakeEnsurer) EnsureBucket(_ context.Context, bucket string) error {
	f.buckets = append(f.buckets, bucket)
	return nil
}
