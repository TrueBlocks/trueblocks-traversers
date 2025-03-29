package traverser

import (
	"strconv"
	"testing"
)

type mockTraverser struct {
	count int
}

func (m *mockTraverser) Traverse(t float64) {
	m.count++
}

func (m *mockTraverser) GetKey(t float64) string {
	return "test"
}

func (m *mockTraverser) Result() string {
	return "Count: " + strconv.Itoa(m.count) // Fix: Use strconv.Itoa to convert int to string
}

func (m *mockTraverser) Name() string {
	return "MockTraverser"
}

func (m *mockTraverser) Sort(array []float64) {
	// No-op for simplicity
}

func TestTraverserInterface(t *testing.T) {
	tr := &mockTraverser{}
	tr.Traverse(1.0)
	tr.Traverse(2.0)

	if tr.count != 2 {
		t.Errorf("Traverse failed: got %d, want 2", tr.count)
	}
	if key := tr.GetKey(1.0); key != "test" {
		t.Errorf("GetKey failed: got %s, want 'test'", key)
	}
	if result := tr.Result(); result != "Count: 2" {
		t.Errorf("Result failed: got %q, want %q", result, "Count: 2")
	}
	if name := tr.Name(); name != "MockTraverser" {
		t.Errorf("Name failed: got %s, want 'MockTraverser'", name)
	}
	// Sort is no-op, so just call it to ensure no panic
	tr.Sort([]float64{1.0, 2.0})
}

