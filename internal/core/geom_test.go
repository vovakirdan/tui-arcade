package core

import "testing"

func TestRectIntersects(t *testing.T) {
	tests := []struct {
		name     string
		a, b     Rect
		expected bool
	}{
		{
			name:     "overlapping rects",
			a:        NewRect(0, 0, 10, 10),
			b:        NewRect(5, 5, 10, 10),
			expected: true,
		},
		{
			name:     "non-overlapping horizontal",
			a:        NewRect(0, 0, 10, 10),
			b:        NewRect(15, 0, 10, 10),
			expected: false,
		},
		{
			name:     "non-overlapping vertical",
			a:        NewRect(0, 0, 10, 10),
			b:        NewRect(0, 15, 10, 10),
			expected: false,
		},
		{
			name:     "adjacent horizontal (no overlap)",
			a:        NewRect(0, 0, 10, 10),
			b:        NewRect(10, 0, 10, 10),
			expected: false,
		},
		{
			name:     "adjacent vertical (no overlap)",
			a:        NewRect(0, 0, 10, 10),
			b:        NewRect(0, 10, 10, 10),
			expected: false,
		},
		{
			name:     "contained rect",
			a:        NewRect(0, 0, 20, 20),
			b:        NewRect(5, 5, 5, 5),
			expected: true,
		},
		{
			name:     "single pixel overlap",
			a:        NewRect(0, 0, 10, 10),
			b:        NewRect(9, 9, 10, 10),
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.a.Intersects(tc.b)
			if result != tc.expected {
				t.Errorf("Intersects() = %v, expected %v", result, tc.expected)
			}
			// Also test symmetry
			resultReverse := tc.b.Intersects(tc.a)
			if resultReverse != tc.expected {
				t.Errorf("Intersects() (reversed) = %v, expected %v", resultReverse, tc.expected)
			}
		})
	}
}

func TestRectContains(t *testing.T) {
	r := NewRect(10, 10, 20, 15)

	tests := []struct {
		name     string
		x, y     int
		expected bool
	}{
		{"inside", 15, 15, true},
		{"top-left corner", 10, 10, true},
		{"bottom-right edge (exclusive)", 30, 25, false},
		{"outside left", 5, 15, false},
		{"outside right", 35, 15, false},
		{"outside top", 15, 5, false},
		{"outside bottom", 15, 30, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := r.Contains(tc.x, tc.y)
			if result != tc.expected {
				t.Errorf("Contains(%d, %d) = %v, expected %v", tc.x, tc.y, result, tc.expected)
			}
		})
	}
}

func TestRectEdges(t *testing.T) {
	r := NewRect(5, 10, 20, 15)

	if r.Right() != 25 {
		t.Errorf("Right() = %d, expected 25", r.Right())
	}
	if r.Bottom() != 25 {
		t.Errorf("Bottom() = %d, expected 25", r.Bottom())
	}

	cx, cy := r.Center()
	if cx != 15 || cy != 17 {
		t.Errorf("Center() = (%d, %d), expected (15, 17)", cx, cy)
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		val, min, max, expected int
	}{
		{5, 0, 10, 5},   // within range
		{-5, 0, 10, 0},  // below min
		{15, 0, 10, 10}, // above max
		{0, 0, 10, 0},   // at min
		{10, 0, 10, 10}, // at max
	}

	for _, tc := range tests {
		result := Clamp(tc.val, tc.min, tc.max)
		if result != tc.expected {
			t.Errorf("Clamp(%d, %d, %d) = %d, expected %d", tc.val, tc.min, tc.max, result, tc.expected)
		}
	}
}

func TestClampF(t *testing.T) {
	tests := []struct {
		val, min, max, expected float64
	}{
		{5.5, 0.0, 10.0, 5.5},
		{-5.5, 0.0, 10.0, 0.0},
		{15.5, 0.0, 10.0, 10.0},
	}

	for _, tc := range tests {
		result := ClampF(tc.val, tc.min, tc.max)
		if result != tc.expected {
			t.Errorf("ClampF(%f, %f, %f) = %f, expected %f", tc.val, tc.min, tc.max, result, tc.expected)
		}
	}
}

func TestMinMax(t *testing.T) {
	if Min(5, 10) != 5 {
		t.Error("Min(5, 10) should be 5")
	}
	if Min(10, 5) != 5 {
		t.Error("Min(10, 5) should be 5")
	}
	if Max(5, 10) != 10 {
		t.Error("Max(5, 10) should be 10")
	}
	if Max(10, 5) != 10 {
		t.Error("Max(10, 5) should be 10")
	}
}

func TestAbs(t *testing.T) {
	if Abs(5) != 5 {
		t.Error("Abs(5) should be 5")
	}
	if Abs(-5) != 5 {
		t.Error("Abs(-5) should be 5")
	}
	if Abs(0) != 0 {
		t.Error("Abs(0) should be 0")
	}
}
