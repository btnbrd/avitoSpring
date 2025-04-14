package integration

import (
	"avitoSpring/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestModeratorRegisterLoginAndCreatePVZ(t *testing.T) {
	username := fmt.Sprintf("moderator_%d@mail.com", time.Now().UnixNano())
	password := "securepassword123"

	t.Run("RegisterModerator", func(t *testing.T) {

		body, _ := json.Marshal(map[string]string{
			"role":     "moderator",
			"email":    username,
			"password": password,
		})
		t.Log(string(body))
		req, err := http.NewRequest(http.MethodPost, baseURL+"/register", bytes.NewReader(body))
		t.Log(req)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode, "Expected status 201 Created")
		t.Log("Registered new moderator")
	})

	var token string
	t.Run("LoginModerator", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"email":    username,
			"password": password,
		})

		req, err := http.NewRequest(http.MethodPost, baseURL+"/login", bytes.NewReader(body))

		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)

		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200 OK")

		var loginResp struct {
			Token string `json:"token"`
		}

		err = json.NewDecoder(resp.Body).Decode(&loginResp)
		require.NoError(t, err)
		require.NotEmpty(t, loginResp.Token)
		token = loginResp.Token
		t.Logf("Logged in as moderator, token: %s", token)
	})

	// Создание ПВЗ
	t.Run("CreatePVZWithToken", func(t *testing.T) {
		body, _ := json.Marshal(models.PVZ{
			City: models.CityKazan,
		})
		req, err := http.NewRequest(http.MethodPost, baseURL+"/pvz", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode, "Expected status 201 Created")

		var pvz models.PVZ
		err = json.NewDecoder(resp.Body).Decode(&pvz)
		require.NoError(t, err)
		require.NotEmpty(t, pvz.ID)
		t.Logf("Created PVZ with ID: %s", pvz.ID)
	})
}
