package viacep

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const url = "https://viacep.com.br"

type Address struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Unidade     string `json:"unidade"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Estado      string `json:"estado"`
	Regiao      string `json:"regiao"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

type ViaCep struct {
}

func New(opts ...func(*ViaCep)) *ViaCep {
	v := &ViaCep{}
	for _, o := range opts {
		o(v)
	}

	return v
}

func (v *ViaCep) Cep(ctx context.Context, cep string) (*Address, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/ws/%s/json/", url, cep), nil)
	if err != nil {
		return nil, err
	}

	c := http.DefaultClient
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("viaCep API returned invalid status code: %d", resp.StatusCode)
	}

	var address Address
	if err = json.NewDecoder(resp.Body).Decode(&address); err != nil {
		return nil, err
	}

	return &address, nil
}

func (v *ViaCep) Addresses(ctx context.Context, uf string, cidade string, logradouro string) ([]Address, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/ws/%s/%s/%s/json/", url, uf, cidade, logradouro), nil)
	if err != nil {
		return nil, err
	}

	c := http.DefaultClient
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("viaCep API returned invalid status code: %d", resp.StatusCode)
	}

	var addresses []Address
	if err = json.NewDecoder(resp.Body).Decode(&addresses); err != nil {
		return nil, err
	}

	return addresses, nil
}
