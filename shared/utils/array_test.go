package utils

import "testing"

func TestUnpackArray(t *testing.T) {
	t.Run("int slice", func(t *testing.T) {
		// Arrange
		in := []int{1, 2, 3}

		// Act
		out := UnpackArray(in)

		// Assert
		if len(out) != len(in) {
			t.Fatalf("len = %d, want %d", len(out), len(in))
		}
		for i, v := range in {
			if out[i] != v {
				t.Errorf("out[%d] = %v, want %v", i, out[i], v)
			}
		}
	})

	t.Run("string slice", func(t *testing.T) {
		in := []string{"a", "b", "c"}
		out := UnpackArray(in)

		if len(out) != len(in) {
			t.Fatalf("len = %d, want %d", len(out), len(in))
		}
		for i, v := range in {
			if out[i] != v {
				t.Errorf("out[%d] = %v, want %v", i, out[i], v)
			}
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		in := []float64{}
		out := UnpackArray(in)
		if len(out) != 0 {
			t.Errorf("want empty slice, got len %d", len(out))
		}
	})
}
