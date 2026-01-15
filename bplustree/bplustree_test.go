package bplustree_test

import (
	"math/rand"
	"testing"

	"github.com/tuananhlai/prototypes/bplustree"
)

func TestInsertGetSingle(t *testing.T) {
	tree := bplustree.New(3)
	tree.Insert(10, "ten")
	v, ok := tree.Get(10)
	if !ok {
		t.Fatalf("expected key to exist")
	}
	if v != "ten" {
		t.Fatalf("expected value %q, got %v", "ten", v)
	}
}

func TestInsertReplace(t *testing.T) {
	tree := bplustree.New(3)
	tree.Insert(5, "a")
	tree.Insert(5, "b")
	v, ok := tree.Get(5)
	if !ok {
		t.Fatalf("expected key to exist")
	}
	if v != "b" {
		t.Fatalf("expected value %q, got %v", "b", v)
	}
}

func TestInsertManyWithSplits(t *testing.T) {
	tree := bplustree.New(3)
	for i := 1; i <= 50; i++ {
		tree.Insert(i, i*10)
	}

	for i := 1; i <= 50; i++ {
		v, ok := tree.Get(i)
		if !ok {
			t.Fatalf("missing key %d", i)
		}
		if v != i*10 {
			t.Fatalf("unexpected value for key %d: %v", i, v)
		}
	}
}

func TestInsertRandomOrder(t *testing.T) {
	tree := bplustree.New(4)
	keys := rand.New(rand.NewSource(42)).Perm(200)
	for _, k := range keys {
		tree.Insert(k, k+1)
	}
	for _, k := range keys {
		v, ok := tree.Get(k)
		if !ok {
			t.Fatalf("missing key %d", k)
		}
		if v != k+1 {
			t.Fatalf("unexpected value for key %d: %v", k, v)
		}
	}
}

func TestGetMissingKey(t *testing.T) {
	tree := bplustree.New(3)
	tree.Insert(1, "one")
	if _, ok := tree.Get(2); ok {
		t.Fatalf("expected missing key to return ok=false")
	}
}
