package service

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	results      map[string]int64
	mu           sync.Mutex
	once         map[string]bool
	instructions []Instruction
}

func NewCalculatorService() *CalculatorService {
	return &CalculatorService{
		results: make(map[string]int64),
		once:    make(map[string]bool),
	}
}

func (s *CalculatorService) Run(ctx context.Context, instructions []Instruction) ([]ResultItem, error) {
	s.instructions = instructions
	var wg sync.WaitGroup
	errChan := make(chan error, 1)
	resultChan := make(chan ResultItem, len(instructions))

	tasks := make(map[string]func() (int64, error))
	printVars := make([]string, 0)
	for _, instr := range instructions {
		instr := instr
		if instr.Type == "calc" {
			name := instr.Var

			if _, exists := tasks[name]; exists {
				return nil, fmt.Errorf("duplicate assignment for variable %s", name)
			}

			tasks[name] = func() (int64, error) {
				log.Printf("Задача %s начата", name)
				defer log.Printf("Задача %s завершена", name)

				s.mu.Lock()
				if s.once[name] {
					s.mu.Unlock()
					return 0, fmt.Errorf("variable %s already assigned", name)
				}
				s.once[name] = true
				s.mu.Unlock()

				time.Sleep(10 * time.Millisecond)

				lVal, err := s.resolveValue(tasks, instr.Left)
				if err != nil {
					return 0, err
				}

				var rVal int64
				if instr.Right != nil {
					rVal, err = s.resolveValue(tasks, instr.Right)
					if err != nil {
						return 0, err
					}
				}

				var res int64
				switch instr.Op {
				case "+":
					res = lVal + rVal
				case "-":
					res = lVal - rVal
				case "*":
					res = lVal * rVal
				default:
					return 0, fmt.Errorf("unsupported operation: %s", instr.Op)
				}

				s.mu.Lock()
				s.results[name] = res
				s.mu.Unlock()

				return res, nil
			}
		} else if instr.Type == "print" {
			printVars = append(printVars, instr.Var)
		}
	}

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

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	output := make([]ResultItem, 0)
	for item := range resultChan {
		output = append(output, item)
	}

	select {
	case err := <-errChan:
		return nil, err
	default:
	}

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
	log.Printf("Получено значение: %+v", val)
	switch v := val.(type) {
	case float64:
		log.Printf("float64 значение: %+v", val)
		if v != float64(int64(v)) {
			return 0, fmt.Errorf("expected integer value, got %v", v)
		}
		return int64(v), nil
	case int64:
		log.Printf("int64 значение: %+v", val)
		return v, nil
	case string:
		log.Printf("string значение: %+v", val)
		s.mu.Lock()
		existingVal, ok := s.results[v]
		if ok {
			s.mu.Unlock()
			return existingVal, nil
		}

		if s.once[v] {
			s.mu.Unlock()
			for i := 0; i < 10; i++ {
				s.mu.Lock()
				if val, ok := s.results[v]; ok {
					s.mu.Unlock()
					return val, nil
				}
				s.mu.Unlock()
				time.Sleep(50 * time.Millisecond)
			}
			return 0, fmt.Errorf("timeout waiting for variable %s", v)
		}

		task, exists := tasks[v]
		if !exists {
			s.mu.Unlock()
			return 0, errors.New("undefined variable: " + v)
		}
		s.mu.Unlock()

		val, err := task()
		if err != nil {
			return 0, err
		}
		return val, nil
	default:
		return 0, errors.New("invalid value type")
	}
}

// func (s *CalculatorService) resolveValue(tasks map[string]func() (int64, error), val interface{}) (int64, error) {
// 	switch v := val.(type) {
// 	case float64:
// 		if v != float64(int64(v)) {
// 			return 0, fmt.Errorf("expected integer value, got %v", v)
// 		}
// 		return int64(v), nil
// 	case int64:
// 		return v, nil
// 	case string:
// 		s.mu.Lock()
// 		alreadyOnce := s.once[v]
// 		existingVal, ok := s.results[v]
// 		if ok {
// 			s.mu.Unlock()
// 			return existingVal, nil
// 		}

// 		if alreadyOnce {
// 			s.mu.Unlock()
// 			return 0, fmt.Errorf("variable %s not ready and already assigned", v)
// 		}

// 		task, exists := tasks[v]
// 		if !exists {
// 			s.mu.Unlock()
// 			return 0, errors.New("undefined variable: " + v)
// 		}

// 		s.mu.Unlock() // разблокируем, чтобы дать другим читать значения

// 		val, err := task()
// 		if err != nil {
// 			return 0, err
// 		}
// 		return val, nil
// 	default:
// 		return 0, errors.New("invalid value type")
// 	}
// }

// func (s *CalculatorService) resolveValue(tasks map[string]func() (int64, error), val interface{}) (int64, error) {
// 	switch v := val.(type) {
// 	case float64:
// 		if v != float64(int64(v)) {
// 			return 0, fmt.Errorf("expected integer value, got %v", v)
// 		}
// 		return int64(v), nil
// 	case int64:
// 		return v, nil
// 	case string:
// 		s.mu.Lock()
// 		existingVal, ok := s.results[v]
// 		alreadyOnce := s.once[v]
// 		s.mu.Unlock()

// 		if ok {
// 			return existingVal, nil
// 		}

// 		if alreadyOnce {
// 			return 0, fmt.Errorf("variable %s not ready and already assigned", v)
// 		}

// 		task, exists := tasks[v]
// 		if !exists {
// 			return 0, errors.New("undefined variable: " + v)
// 		}

// 		val, err := task()
// 		if err != nil {
// 			return 0, err
// 		}
// 		return val, nil
// 	default:
// 		return 0, errors.New("invalid value type")
// 	}
// }
