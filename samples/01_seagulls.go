// Chapter 01: Seagulls — Why FP matters
// Demonstrates the problem with mutable state and OOP
package main

import "fmt"

type Flock struct {
	Seagulls int
}

func MakeFlock(n int) Flock {
	return Flock{Seagulls: n}
}

func (f *Flock) Conjoin(other *Flock) *Flock {
	f.Seagulls += other.Seagulls
	return f
}

func (f *Flock) Breed(other *Flock) *Flock {
	f.Seagulls = f.Seagulls * other.Seagulls
	return f
}

func main() {
	fmt.Println("=== Chapter 01: The Seagull Problem ===")
	fmt.Println()

	// OOP style - mutable, surprising results
	flockA := MakeFlock(4)
	flockB := MakeFlock(2)
	flockC := MakeFlock(0)

	result := flockA.Conjoin(&flockC).Breed(&flockB).Conjoin(flockA.Breed(&flockB)).Seagulls
	fmt.Println("OOP mutable result:", result)
	fmt.Println("Expected: 16, Got:", result, "(mutation causes bugs!)")
	fmt.Println()

	// FP style - using simple functions, no mutation
	add := func(a, b int) int { return a + b }
	multiply := func(a, b int) int { return a * b }

	a, b, c := 4, 2, 0
	fpResult := add(multiply(add(a, c), b), multiply(a, b))
	fmt.Println("FP pure result:", fpResult)
	fmt.Println("Correct! Pure functions give predictable results.")
}
