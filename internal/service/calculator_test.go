package service

import (
	"context"
	"strings"
	"testing"
)

func TestCalculator_Run_SimpleCalc(t *testing.T) {
	// t.Parallel()
	s := NewCalculatorService()

	instructions := []Instruction{
		{
			Type:  "calc",
			Op:    "+",
			Var:   "x",
			Left:  int64(1),
			Right: int64(2),
		},
		{
			Type: "print",
			Var:  "x",
		},
	}

	results, err := s.Run(context.Background(), instructions)
	if err != nil {
		t.Errorf("ожидается успех, получена ошибка: %v", err)
	}

	if len(results) != 1 || results[0].Value != 3 {
		t.Errorf("ожидается результат x=3, получено %+v", results)
	}
}

func TestCalculator_Run_VariableDependency(t *testing.T) {

	// t.Parallel()
	s := NewCalculatorService()

	instructions := []Instruction{
		{
			Type:  "calc",
			Op:    "+",
			Var:   "x",
			Left:  int64(10),
			Right: int64(2),
		},
		{
			Type:  "calc",
			Op:    "-",
			Var:   "y",
			Left:  "x",
			Right: int64(3),
		},
		{
			Type: "print",
			Var:  "y",
		},
	}

	results, err := s.Run(context.Background(), instructions)
	if err != nil {
		t.Errorf("ожидается успех, получена ошибка: %v", err)
	}

	if len(results) != 1 || results[0].Value != 9 {
		t.Errorf("ожидается y=9, получено %+v", results)
	}
}

func TestCalculator_Run_ReassignVariable(t *testing.T) {
	// t.Parallel()
	s := NewCalculatorService()

	instructions := []Instruction{
		{
			Type:  "calc",
			Op:    "+",
			Var:   "x",
			Left:  int64(1),
			Right: int64(2),
		},
		{
			Type:  "calc",
			Op:    "*",
			Var:   "x",
			Left:  int64(3),
			Right: int64(4),
		},
		{
			Type: "print",
			Var:  "x",
		},
	}

	_, err := s.Run(context.Background(), instructions)
	if err == nil {
		t.Fatal("ожидается ошибка повторного назначения, но её нет")
	}

	expectedError := "duplicate assignment for variable x"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("ожидается ошибка содержащая %q, получено: %v", expectedError, err)
	}
}

func TestCalculator_Run_UndefinedVariable(t *testing.T) {
	// t.Parallel()
	s := NewCalculatorService()

	instructions := []Instruction{
		{
			Type:  "calc",
			Op:    "+",
			Var:   "y",
			Left:  "x",
			Right: int64(2),
		},
		{
			Type: "print",
			Var:  "y",
		},
	}

	_, err := s.Run(context.Background(), instructions)
	if err == nil {
		t.Error("ожидается ошибка 'undefined variable', но её нет")
		return
	}

	if err.Error() != "undefined variable: x" {
		t.Errorf("ожидается 'undefined variable: x', получено: %v", err)
	}
}

func TestCalculator_Run_ComplexGraph(t *testing.T) {
	s := NewCalculatorService()

	instructions := []Instruction{
		{Type: "calc", Op: "+", Var: "x", Left: int64(10), Right: int64(2)},
		{Type: "calc", Op: "*", Var: "y", Left: "x", Right: int64(5)},
		{Type: "calc", Op: "-", Var: "q", Left: "y", Right: int64(20)},
		{Type: "calc", Op: "+", Var: "unusedA", Left: "y", Right: int64(100)},
		{Type: "calc", Op: "*", Var: "unusedB", Left: "unusedA", Right: "y"},
		{Type: "print", Var: "q"},
		{Type: "calc", Op: "-", Var: "z", Left: "x", Right: int64(15)},
		{Type: "print", Var: "z"},
		{Type: "calc", Op: "+", Var: "ignoreC", Left: "z", Right: "y"},
		{Type: "print", Var: "x"},
	}

	results, err := s.Run(context.Background(), instructions)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	expected := map[string]int64{
		"q": 40,
		"z": -3,
		"x": 12,
	}

	if len(results) != len(expected) {
		t.Fatalf("ожидается %d результатов, получено %d", len(expected), len(results))
	}

	for _, item := range results {
		val, ok := expected[item.Var]
		if !ok {
			t.Errorf("непредвиденная переменная: %s", item.Var)
			continue
		}
		if item.Value != val {
			t.Errorf("для %s ожидается %d, получено %d", item.Var, val, item.Value)
		}
	}
}

// func TestCalculator_Run_ComplexGraph(t *testing.T) {
// 	// t.Parallel()
// 	s := NewCalculatorService()

// 	instructions := []Instruction{
// 		{Type: "calc", Op: "+", Var: "x", Left: int64(10), Right: int64(2)},
// 		{Type: "calc", Op: "*", Var: "y", Left: "x", Right: int64(5)},
// 		{Type: "calc", Op: "-", Var: "q", Left: "y", Right: int64(20)},
// 		{Type: "calc", Op: "+", Var: "unusedA", Left: "y", Right: int64(100)},
// 		{Type: "calc", Op: "*", Var: "unusedB", Left: "unusedA", Right: "y"},
// 		{Type: "print", Var: "q"},
// 		{Type: "calc", Op: "-", Var: "z", Left: "x", Right: int64(15)},
// 		{Type: "print", Var: "z"},
// 		{Type: "calc", Op: "+", Var: "ignoreC", Left: "z", Right: "y"},
// 		{Type: "print", Var: "x"},
// 	}

// 	results, err := s.Run(context.Background(), instructions)
// 	if err != nil {
// 		t.Errorf("ожидается успех, получена ошибка: %v", err)
// 	}

// 	expected := map[string]int64{
// 		"q": 30,
// 		"z": -3,
// 		"x": 12,
// 	}

// 	for _, item := range results {
// 		val, ok := expected[item.Var]
// 		if !ok {
// 			t.Errorf("непредвиденная переменная: %s", item.Var)
// 			continue
// 		}
// 		if item.Value != val {
// 			t.Errorf("для %s ожидается %d, получено %d", item.Var, val, item.Value)
// 		}
// 	}
// }
