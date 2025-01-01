package viacep

import (
	"context"
	"fmt"
)

const urlBase = "https://viacep.com.br"

type Service interface {
	// Cep retrieves the address information for a given CEP (postal code).
	//
	// Parameters:
	//   - ctx: The context to manage the request lifecycle, such as timeouts or cancellations.
	//   - cep: The postal code (CEP) for which the address information will be retrieved.
	//
	// Returns:
	//   - *Address: A pointer to the Address object with the address data.
	//   - error: If an error occurs during the request, it will be returned. Otherwise, nil will be returned.
	Cep(ctx context.Context, cep string) (*Address, error)

	// Addresses retrieves a list of addresses based on the provided parameters: state (uf), city (cidade), and street (logradouro).
	//
	// Parameters:
	//   - ctx: The context to manage the request lifecycle, such as timeouts or cancellations.
	//   - uf: The federative unit (state) for which the address search will be conducted.
	//   - cidade: The name of the city for which the address search will be conducted.
	//   - logradouro: The name of the street or address for which the address search will be conducted.
	//
	// Returns:
	//   - []Address: A list of addresses found.
	//   - error: If an error occurs during the request, it will be returned. Otherwise, nil will be returned.
	Addresses(ctx context.Context, uf string, cidade string, logradouro string) ([]Address, error)
}

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
	httpClient Http
	cache      Cache
}

func New(httpClient Http) *ViaCep {
	return &ViaCep{
		httpClient: httpClient,
		cache:      newMemoryCache(),
	}
}

func (v *ViaCep) Cep(ctx context.Context, cep string) (*Address, error) {
	key := cacheKey(cep)

	var address Address
	if found := v.cache.Get(ctx, key, &address); found {
		return &address, nil
	}

	url := fmt.Sprintf("%s/ws/%s/json/", urlBase, cep)
	if err := v.httpClient.Get(ctx, url, &address); err != nil {
		return nil, err
	}

	_ = v.cache.Set(ctx, key, address, cacheTTL)
	return &address, nil
}

func (v *ViaCep) Addresses(ctx context.Context, uf string, cidade string, logradouro string) ([]Address, error) {
	key := cacheKey(uf, cidade, logradouro)

	var addresses []Address
	if found := v.cache.Get(ctx, key, &addresses); found {
		return addresses, nil
	}

	url := fmt.Sprintf("%s/ws/%s/%s/%s/json/", urlBase, uf, cidade, logradouro)
	if err := v.httpClient.Get(ctx, url, &addresses); err != nil {
		return nil, err
	}

	_ = v.cache.Set(ctx, key, addresses, cacheTTL)
	return addresses, nil
}
