package viacep

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestViaCep_Client_Cep(t *testing.T) {
	t.Run("integration", func(t *testing.T) {
		if testing.Short() {
			t.Log("integration testing skipped")
			t.Skip()
		}

		c := New(NewHttpClient(1))
		address, err := c.Cep(context.Background(), "01001000")
		assert.NoError(t, err)

		expected := Address{
			Cep:         "01001-000",
			Logradouro:  "Praça da Sé",
			Complemento: "lado ímpar",
			Unidade:     "",
			Bairro:      "Sé",
			Localidade:  "São Paulo",
			Uf:          "SP",
			Estado:      "São Paulo",
			Regiao:      "Sudeste",
			Ibge:        "3550308",
			Gia:         "1004",
			Ddd:         "11",
			Siafi:       "7107",
		}

		assert.Equal(t, &expected, address)
	})
}

func TestViaCep_Client_Addresses(t *testing.T) {
	t.Run("integration", func(t *testing.T) {
		if testing.Short() {
			t.Log("integration testing skipped")
			t.Skip()
		}

		c := New(NewHttpClient(1))
		addresses, err := c.Addresses(context.Background(), "RS", "Porto Alegre", "Domingos+José")
		assert.NoError(t, err)

		expected := []Address{
			{Cep: "91790-072", Logradouro: "Rua Domingos José Poli", Complemento: "", Unidade: "", Bairro: "Restinga", Localidade: "Porto Alegre", Uf: "RS", Estado: "Rio Grande do Sul", Regiao: "Sul", Ibge: "4314902", Gia: "", Ddd: "51", Siafi: "8801"},
			{Cep: "91910-420", Logradouro: "Rua José Domingos Varella", Complemento: "", Unidade: "", Bairro: "Cavalhada", Localidade: "Porto Alegre", Uf: "RS", Estado: "Rio Grande do Sul", Regiao: "Sul", Ibge: "4314902", Gia: "", Ddd: "51", Siafi: "8801"},
			{Cep: "90420-200", Logradouro: "Rua Domingos José de Almeida", Complemento: "", Unidade: "", Bairro: "Rio Branco", Localidade: "Porto Alegre", Uf: "RS", Estado: "Rio Grande do Sul", Regiao: "Sul", Ibge: "4314902", Gia: "", Ddd: "51", Siafi: "8801"},
		}

		assert.Equal(t, expected, addresses)
	})
}
