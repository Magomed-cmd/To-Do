package main

import (
	"fmt"
)

// sum возвращает сумму двух целых чисел
func sum(a int, b int) int {
	return a + b
}

func main() {
	a := 5
	b := 7
	result := sum(a, b)
	fmt.Printf("Сумма %d и %d равна %d\n", a, b, result)
}
