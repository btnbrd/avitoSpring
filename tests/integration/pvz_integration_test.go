package integration

import (
	"avitoSpring/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL = "http://localhost:8080"
)

var client = &http.Client{Timeout: 10 * time.Second}

func TestPVZIntegration(t *testing.T) {

	var pvzID, receptionID string
	var authHeaderModerator, authHeaderEmployee string

	t.Run("GetAuthTokens", func(t *testing.T) {
		token, err := getAuthToken("moderator")
		require.NoError(t, err, "Failed to get moderator auth token")
		authHeaderModerator = fmt.Sprintf("Bearer %s", token)
		t.Logf("Got moderator token")

		tokenEmployee, err := getAuthToken("employee")
		require.NoError(t, err, "Failed to get employee auth token")
		authHeaderEmployee = fmt.Sprintf("Bearer %s", tokenEmployee)
		t.Logf("Got employee token")
	})

	t.Run("CreatePVZ", func(t *testing.T) {
		var err error
		pvzID, err = createPVZ(authHeaderModerator)
		require.NoError(t, err, "Failed to create PVZ")
		t.Logf("Created PVZ with ID: %s", pvzID)
	})

	// Подтест 2: Создание приёмки
	t.Run("CreateReception", func(t *testing.T) {
		var err error
		receptionID, err = createReception(authHeaderEmployee, pvzID)
		require.NoError(t, err, "Failed to create reception")
		t.Logf("Created reception with ID: %s", receptionID)
	})

	// Подтест 3: Добавление 50 товаров
	t.Run("Create50Products", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			productType := models.ProductTypeFootwear
			switch rand.Intn(3) {
			case 0:
				productType = models.ProductTypeElectronics
			case 1:
				productType = models.ProductTypeClothing
			default:
				productType = models.ProductTypeFootwear
			}
			productID, err := createProduct(authHeaderEmployee, pvzID, productType)
			require.NoError(t, err, "Failed to create product #%d", i+1)
			t.Logf("Created product #%d with ID: %s", i+1, productID)
		}
	})

	t.Run("CloseReception", func(t *testing.T) {
		closedReception, err := closeLastReception(authHeaderEmployee, pvzID)
		require.NoError(t, err, "Failed to close reception")
		assert.Equal(t, models.ReceptionStatusClose, closedReception.Status, "Reception should be closed")
		assert.Equal(t, receptionID, closedReception.ID, "Closed reception ID should match")
		t.Logf("Closed reception with ID: %s", closedReception.ID)
	})

}

func getAuthToken(role string) (string, error) {
	body, _ := json.Marshal(map[string]string{"role": role})
	req, err := http.NewRequest(http.MethodPost, baseURL+"/dummyLogin", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}
	return response.Token, nil
}

func createPVZ(authHeader string) (string, error) {
	var city models.City
	switch rand.Intn(3) {
	case 0:
		city = models.CityMoscow
	case 1:
		city = models.CityKazan
	default:
		city = models.CitySaintPetersburg
	}
	body, _ := json.Marshal(models.PVZ{
		City: city,
		//RegistrationDate: time.Now().UTC().Format(time.RFC3339),
	})
	req, err := http.NewRequest(http.MethodPost, baseURL+"/pvz", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errorResp models.Error
		json.NewDecoder(resp.Body).Decode(&errorResp)
		return "", fmt.Errorf("unexpected status code: %d, message: %s", resp.StatusCode, errorResp.Message)
	}

	var pvz models.PVZ
	if err := json.NewDecoder(resp.Body).Decode(&pvz); err != nil {
		return "", err
	}
	return pvz.ID, nil
}

func createReception(authHeader, pvzID string) (string, error) {
	body, _ := json.Marshal(map[string]string{"pvzId": pvzID})
	req, err := http.NewRequest(http.MethodPost, baseURL+"/receptions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errorResp models.Error
		json.NewDecoder(resp.Body).Decode(&errorResp)
		return "", fmt.Errorf("unexpected status code: %d, message: %s", resp.StatusCode, errorResp.Message)
	}

	var reception models.Reception
	if err := json.NewDecoder(resp.Body).Decode(&reception); err != nil {
		return "", err
	}
	return reception.ID, nil
}

func createProduct(authHeader, pvzID string, productType models.ProductType) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"type":  string(productType),
		"pvzId": pvzID,
	})
	req, err := http.NewRequest(http.MethodPost, baseURL+"/products", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errorResp models.Error
		json.NewDecoder(resp.Body).Decode(&errorResp)
		return "", fmt.Errorf("unexpected status code: %d, message: %s", resp.StatusCode, errorResp.Message)
	}

	var product models.Product
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return "", err
	}
	return product.ID, nil
}

func closeLastReception(authHeader, pvzID string) (*models.Reception, error) {
	req, err := http.NewRequest(http.MethodPost, baseURL+"/pvz/"+pvzID+"/close_last_reception", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authHeader)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp models.Error
		json.NewDecoder(resp.Body).Decode(&errorResp)
		return nil, fmt.Errorf("unexpected status code: %d, message: %s", resp.StatusCode, errorResp.Message)
	}

	var reception models.Reception
	if err := json.NewDecoder(resp.Body).Decode(&reception); err != nil {
		return nil, err
	}
	return &reception, nil
}
