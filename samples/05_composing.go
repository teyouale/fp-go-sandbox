// Chapter 05: Composing — Function composition with fp-go
package main

import (
	"fmt"
	"regexp"
	"strings"

	A "github.com/IBM/fp-go/array"
	F "github.com/IBM/fp-go/function"
	I "github.com/IBM/fp-go/number/integer"
	O "github.com/IBM/fp-go/option"
	"github.com/IBM/fp-go/ord"
	S "github.com/IBM/fp-go/string"
	N "github.com/IBM/fp-go/number"
)

type Car struct {
	Name        string
	Horsepower  int
	DollarValue float32
	InStock     bool
}

func (c Car) getInStock() bool    { return c.InStock }
func (c Car) getDollarValue() float32 { return c.DollarValue }
func (c Car) getHorsepower() int  { return c.Horsepower }
func (c Car) getName() string     { return c.Name }

func average(val []float32) float32 {
	return F.Pipe2(
		val,
		A.Fold(N.MonoidSum[float32]()),
		N.Div(float32(len(val))),
	)
}

var (
	ToLower  = strings.ToLower
	ToUpper  = strings.ToUpper
	Exclaim  = S.Format[string]("%s!")
	Shout    = F.Flow2(ToUpper, Exclaim)

	Replace = func(search *regexp.Regexp) func(replace string) func(s string) string {
		return func(replace string) func(s string) string {
			return func(s string) string {
				return search.ReplaceAllString(s, replace)
			}
		}
	}

	Split = F.Curry2(F.Bind3of3((*regexp.Regexp).Split)(-1))

	Dasherize = F.Flow4(
		Replace(regexp.MustCompile(`\s{2,}`))(" "),
		Split(regexp.MustCompile(` `)),
		A.Map(ToLower),
		A.Intercalate(S.Monoid)("-"),
	)

	Cars = A.From(
		Car{"Ferrari FF", 660, 700000, true},
		Car{"Spyker C12 Zagato", 650, 648000, false},
		Car{"Jaguar XKR-S", 550, 132000, true},
		Car{"Audi R8", 525, 114200, false},
		Car{"Aston Martin One-77", 750, 1850000, true},
		Car{"Pagani Huayra", 700, 1300000, false},
	)
)

func main() {
	fmt.Println("=== Chapter 05: Composing ===")
	fmt.Println()

	// Shout: compose ToUpper and Exclaim
	fmt.Println("Shout(\"send in the clowns\"):")
	fmt.Println(" ", Shout("send in the clowns"))
	fmt.Println()

	// Dasherize: a multi-step composition
	fmt.Println("Dasherize(\"The world is a vampire\"):")
	fmt.Println(" ", Dasherize("The world is a vampire"))
	fmt.Println()

	// Pipe: explicit left-to-right piping
	output := F.Pipe2("send in the clowns", ToUpper, Exclaim)
	fmt.Println("Pipe(\"send in the clowns\", ToUpper, Exclaim):")
	fmt.Println(" ", output)
	fmt.Println()

	// IsLastInStock: check if last car is in stock
	IsLastInStock := F.Flow2(A.Last[Car], O.Map(Car.getInStock))
	fmt.Println("IsLastInStock(first 3 cars):", IsLastInStock(Cars[0:3]))
	fmt.Println("IsLastInStock(last 3 cars): ", IsLastInStock(Cars[3:]))
	fmt.Println()

	// AverageDollarValue
	averageDollarValue := F.Flow2(A.Map(Car.getDollarValue), average)
	fmt.Println("averageDollarValue(all cars):", averageDollarValue(Cars))
	fmt.Println()

	// FastestCar
	ordByHorsepower := ord.Contramap(Car.getHorsepower)(I.Ord)
	fastestCar := F.Flow3(
		A.Sort(ordByHorsepower),
		A.Last[Car],
		O.Map(F.Flow2(Car.getName, S.Format[string]("%s is the fastest"))),
	)
	fmt.Println("fastestCar:", fastestCar(Cars))
}
