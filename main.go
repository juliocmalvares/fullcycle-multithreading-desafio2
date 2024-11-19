package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type BrasilAPIResponse struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
}

type ViaCEPResponse struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Por favor, forneça o CEP como argumento.")
		return
	}
	cep := os.Args[1]

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	type Result struct {
		APIName string
		Data    interface{}
		Error   error
	}
	resultChan := make(chan Result, 2)

	go func() {
		data, err := fetchFromBrasilAPI(ctx, cep)
		resultChan <- Result{APIName: "BrasilAPI", Data: data, Error: err}
	}()

	go func() {
		data, err := fetchFromViaCEP(ctx, cep)
		resultChan <- Result{APIName: "ViaCEP", Data: data, Error: err}
	}()

	var firstError error
	for i := 0; i < 2; i++ {
		select {
		case res := <-resultChan:
			if res.Error == nil {
				fmt.Printf("Resposta recebida da API %s:\n", res.APIName)
				fmt.Printf("%+v\n", res.Data)
				return
			} else {
				if firstError == nil {
					firstError = fmt.Errorf("erro ao buscar dados da API %s: %v", res.APIName, res.Error)
				} else {
					firstError = fmt.Errorf("%v\nerro ao buscar dados da API %s: %v", firstError, res.APIName, res.Error)
				}
			}
		case <-ctx.Done():
			fmt.Println("Tempo limite de 1 segundo excedido. Erro de timeout.")
			return
		}
	}

	if firstError != nil {
		fmt.Println(firstError)
	}
}

func fetchFromBrasilAPI(ctx context.Context, cep string) (BrasilAPIResponse, error) {
	var result BrasilAPIResponse
	req, err := http.NewRequestWithContext(ctx, "GET", "https://brasilapi.com.br/api/cep/v1/"+cep, nil)
	if err != nil {
		return result, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(body, &result)
	return result, err
}

func fetchFromViaCEP(ctx context.Context, cep string) (ViaCEPResponse, error) {
	var result ViaCEPResponse
	req, err := http.NewRequestWithContext(ctx, "GET", "http://viacep.com.br/ws/"+cep+"/json/", nil)
	if err != nil {
		return result, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(body, &result)
	return result, err
}
