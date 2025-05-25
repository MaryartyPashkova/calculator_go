package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Instruction struct {
	Type  string      `json:"type"`
	Op    string      `json:"op,omitempty"`
	Var   string      `json:"var"`
	Left  interface{} `json:"left"`
	Right interface{} `json:"right,omitempty"`
}

type ResultItem struct {
	Var   string `json:"var"`
	Value int64  `json:"value"`
}

type CalculatorService struct {
	results map[string]int64
	mu      sync.Mutex
	once    map[string]bool
}

func NewCalculatorService() *CalculatorService {
	return &CalculatorService{
		results: make(map[string]int64),
		once:    make(map[string]bool),
	}
}

func (s *CalculatorService) Run(ctx context.Context, instructions []Instruction) ([]ResultItem, error) {
	var wg sync.WaitGroup
	errChan := make(chan error, 1)
	resultChan := make(chan ResultItem, len(instructions))

	tasks := make(map[string]func() (int64, error))
	printVars := make([]string, 0)

	// Разделяем calc и print
	for _, instr := range instructions {
		if instr.Type == "calc" {
			name := instr.Var
			left := instr.Left
			right := instr.Right
			op := instr.Op

			leftCopy := left
			rightCopy := right

			tasks[name] = func() (int64, error) {
				time.Sleep(50 * time.Millisecond)

				lVal, err := s.resolveValue(tasks, leftCopy)
				if err != nil {
					return 0, err
				}

				var rVal int64
				if rightCopy != nil {
					rVal, err = s.resolveValue(tasks, rightCopy)
					if err != nil {
						return 0, err
					}
				}

				var res int64
				switch op {
				case "+":
					res = lVal + rVal
				case "-":
					res = lVal - rVal
				case "*":
					res = lVal * rVal
				default:
					return 0, fmt.Errorf("unsupported operation: %s", op)
				}

				s.mu.Lock()
				if s.once[name] {
					s.mu.Unlock()
					return 0, fmt.Errorf("variable %s already assigned", name)
				}
				s.results[name] = res
				s.once[name] = true
				s.mu.Unlock()

				return res, nil
			}
		} else if instr.Type == "print" {
			printVars = append(printVars, instr.Var)
		}
	}

	// Запуск задач параллельно
	for name, task := range tasks {
		wg.Add(1)
		go func(n string, t func() (int64, error)) {
			val, err := t()
			if err != nil {
				select {
				case errChan <- err:
				default:
				}
			} else {
				resultChan <- ResultItem{Var: n, Value: val}
			}
			wg.Done()
		}(name, task)
	}

	// Ждём завершения
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Собираем результаты
	output := make([]ResultItem, 0)
	for item := range resultChan {
		output = append(output, item)
	}

	select {
	case err := <-errChan:
		return nil, err
	default:
	}

	// Фильтруем по printVars
	finalOutput := make([]ResultItem, 0)
	for _, varName := range printVars {
		val, ok := s.results[varName]
		if !ok {
			continue
		}
		finalOutput = append(finalOutput, ResultItem{Var: varName, Value: val})
	}

	return finalOutput, nil
}

func (s *CalculatorService) resolveValue(tasks map[string]func() (int64, error), val interface{}) (int64, error) {
	switch v := val.(type) {
	case float64:
		if v != float64(int64(v)) {
			return 0, fmt.Errorf("expected integer value, got %v", v)
		}
		return int64(v), nil
	case int64:
		return v, nil
	case string:
		task, exists := tasks[v]
		if !exists {
			return 0, errors.New("undefined variable: " + v)
		}
		return task()
	default:
		return 0, errors.New("invalid value type 8")
	}
}
