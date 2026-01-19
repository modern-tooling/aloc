package main

import "testing"

func TestAdd(t *testing.T) {
    result := Add(1, 2)
    if result != 3 {
        t.Errorf("Add(1, 2) = %d; want 3", result)
    }
}

func TestSubtract(t *testing.T) {
    result := Subtract(5, 3)
    if result != 2 {
        t.Errorf("Subtract(5, 3) = %d; want 2", result)
    }
}

func TestMultiply(t *testing.T) {
    result := Multiply(3, 4)
    if result != 12 {
        t.Errorf("Multiply(3, 4) = %d; want 12", result)
    }
}
