// Chapter 04: Currying — Partial application with fp-go
package main

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	A "github.com/IBM/fp-go/array"
	F "github.com/IBM/fp-go/function"
	N "github.com/IBM/fp-go/number"
	S "github.com/IBM/fp-go/string"
)

var (
	Matches = F.Curry2((*regexp.Regexp).MatchString)
	Split   = F.Curry2(F.Bind3of3((*regexp.Regexp).Split)(-1))

	Add      = N.Add[int]
	ToLower  = strings.ToLower
	ToUpper  = strings.ToUpper
	Concat   = F.Curry2(S.Monoid.Concat)
)

func main() {
	fmt.Println("=== Chapter 04: Currying ===")
	fmt.Println()

	// Solution A: words splits a string into words
	words := Split(regexp.MustCompile(` `))
	fmt.Println("words(\"Jingle bells Batman smells\"):")
	fmt.Println(" ", words("Jingle bells Batman smells"))
	fmt.Println()

	// Solution B: filterQs filters strings containing 'q'
	filterQs := A.Filter(Matches(regexp.MustCompile(`q`)))
	fmt.Println("filterQs([\"quick\", \"camels\", \"quarry\", \"over\", \"quails\"]):")
	fmt.Println(" ", filterQs(A.From("quick", "camels", "quarry", "over", "quails")))
	fmt.Println()

	// Solution C: max finds the maximum number
	keepHighest := N.Max[int]
	max := A.Reduce(keepHighest, math.MinInt)
	fmt.Println("max([323, 523, 554, 123, 5234]):")
	fmt.Println(" ", max(A.From(323, 523, 554, 123, 5234)))
}
