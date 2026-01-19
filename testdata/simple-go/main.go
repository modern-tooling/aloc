package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
    result := Add(1, 2)
    fmt.Printf("1 + 2 = %d\n", result)
}

func Add(a, b int) int {
    return a + b
}

func Subtract(a, b int) int {
    return a - b
}

func Multiply(a, b int) int {
    return a * b
}
