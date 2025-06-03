package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"calculator/internal/service"
)

// func startHTTPServer() {

// 	http.HandleFunc("/calculate", func(w http.ResponseWriter, r *http.Request) {
// 		var instructions []service.Instruction
// 		log.Printf("Получен запрос: %+v", instructions)
// 		log.Printf("Получено %d инструкций", len(instructions))
// 		for i, instr := range instructions {
// 			log.Printf("Инструкция %d: %+v", i, instr)
// 		}
// 		if err := json.NewDecoder(r.Body).Decode(&instructions); err != nil {
// 			http.Error(w, "Invalid request body", http.StatusBadRequest)
// 			return
// 		}

// 		calc := service.NewCalculatorService()
// 		results, err := calc.Run(context.Background(), instructions)
// 		log.Print("Какие-то данные:", results, err)

// 		if err != nil {
// 			http.Error(w, fmt.Sprintf("Execution error: %v", err), http.StatusInternalServerError)
// 			return
// 		}

// 		response := struct {
// 			Items []service.ResultItem `json:"items"`
// 		}{Items: results}
// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusOK)
// 		json.NewEncoder(w).Encode(response)
// 	})

// 	log.Println("HTTP сервер запущен на :8081")
// 	log.Fatal(http.ListenAndServe(":8081", nil))
// }

func startHTTPServer() {
	http.HandleFunc("/calculate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var instructions []service.Instruction
		if err := json.NewDecoder(r.Body).Decode(&instructions); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			log.Printf("Ошибка парсинга: %v", err)
			return
		}

		log.Printf("Получено %d инструкций", len(instructions))
		for i, instr := range instructions {
			log.Printf("Инструкция %d: %+v", i, instr)
		}

		calc := service.NewCalculatorService()
		results, err := calc.Run(context.Background(), instructions)
		if err != nil {
			http.Error(w, fmt.Sprintf("Execution error: %v", err), http.StatusInternalServerError)
			return
		}

		response := struct {
			Items []service.ResultItem `json:"items"`
		}{Items: results}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	log.Println("HTTP сервер запущен на :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		startHTTPServer()
	}()

	go func() {
		defer wg.Done()
		startGRPCServer()
	}()

	wg.Wait()
}
