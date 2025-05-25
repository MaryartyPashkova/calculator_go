package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "calculator/internal/service"
)

func main() {
    input := `[{"type":"calc","op":"+","var":"x","left":1,"right":2},{"type":"print","var":"x"}]`

    var instructions []service.Instruction
    err := json.Unmarshal([]byte(input), &instructions)
    if err != nil {
        log.Fatalf("Ошибка парсинга JSON: %v", err)
    }

    calc := service.NewCalculatorService()
    result, err := calc.Run(context.Background(), instructions)
    if err != nil {
        log.Fatalf("Ошибка выполнения: %v", err)
    }

    jsonResult, _ := json.MarshalIndent(result, "", "  ")
    fmt.Println(string(jsonResult))
}