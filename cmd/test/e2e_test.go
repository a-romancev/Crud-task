package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/a-romancev/crud_task/auth"
	"github.com/a-romancev/crud_task/company"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	addr = "localhost:9999"
	// nolint: gosec
	secretKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIB8fmVWhMdAo/UkDNN4UGo8PYwKxz/lN7nilmYa2KEkboAoGCCqGSM49
AwEHoUQDQgAETrMd0Br7GOpE7US1jJ7LbL0L8vIi3NxRxnXhOxDWaAhd4MxdF17f
AY5OGjJpPdWJ8TDMQH7Es98SAB9pVRVZhg==
-----END EC PRIVATE KEY-----`
)

type Client struct {
	client *http.Client
	sk     *auth.SecretKey
	userID uuid.UUID
}

func NewClient(userID uuid.UUID, client *http.Client) *Client {
	sk, _ := auth.NewSecretKey(secretKey)
	return &Client{
		client: client,
		sk:     sk,
		userID: userID,
	}
}

func (c *Client) Do(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	token, _ := c.sk.Sign(auth.NewAPIClaims(c.userID))
	request.Header.Set("Authorization", "Bearer "+token)
	return c.client.Do(request)
}

func TestAPI(t *testing.T) {
	t.Run("Create company", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			client := NewClient(uuid.New(), http.DefaultClient)

			employees := 1
			reg := true
			tp := "Corporations"
			body, _ := json.Marshal(company.Company{
				Name:         uuid.NewString()[:10],
				EmployeesNum: &employees,
				Registered:   &reg,
				Type:         &tp,
			})

			resp, err := client.Do(ctx, http.MethodPost, fmt.Sprintf("http://%s/v1/companies", addr), bytes.NewReader(body))
			require.NoError(t, err)
			presCmp, err := io.ReadAll(resp.Body)
			c := company.Company{}
			err = json.Unmarshal(presCmp, &c)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
			resp, err = client.Do(ctx, http.MethodGet, fmt.Sprintf("http://%s/v1/companies/%s", addr, c.ID), bytes.NewReader(body))
			require.NoError(t, err)
			gresCmp, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, presCmp, gresCmp)
		})
	})
	t.Run("Update company", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			client := NewClient(uuid.New(), http.DefaultClient)

			employees := 1
			reg := true
			tp := "Corporations"
			body, _ := json.Marshal(company.Company{
				Name:         uuid.NewString()[:10],
				EmployeesNum: &employees,
				Registered:   &reg,
				Type:         &tp,
			})

			resp, err := client.Do(ctx, http.MethodPost, fmt.Sprintf("http://%s/v1/companies", addr), bytes.NewReader(body))
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
			presCmp, err := io.ReadAll(resp.Body)
			c := company.Company{}
			err = json.Unmarshal(presCmp, &c)
			require.NoError(t, err)

			body, _ = json.Marshal(company.Company{
				Name:         uuid.NewString()[:10] + "2",
				EmployeesNum: &employees,
				Registered:   &reg,
				Type:         &tp,
			})
			resp, err = client.Do(ctx, http.MethodPut, fmt.Sprintf("http://%s/v1/companies/%s", addr, c.ID), bytes.NewReader(body))
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		})
	})
	t.Run("Delete company", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			client := NewClient(uuid.New(), http.DefaultClient)

			employees := 1
			reg := true
			tp := "Corporations"
			body, _ := json.Marshal(company.Company{
				Name:         uuid.NewString()[:10],
				EmployeesNum: &employees,
				Registered:   &reg,
				Type:         &tp,
			})

			resp, err := client.Do(ctx, http.MethodPost, fmt.Sprintf("http://%s/v1/companies", addr), bytes.NewReader(body))
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
			presCmp, err := io.ReadAll(resp.Body)
			c := company.Company{}
			err = json.Unmarshal(presCmp, &c)
			require.NoError(t, err)

			body, _ = json.Marshal(company.Company{
				Name:         uuid.NewString()[:10] + "2",
				EmployeesNum: &employees,
				Registered:   &reg,
				Type:         &tp,
			})
			resp, err = client.Do(ctx, http.MethodDelete, fmt.Sprintf("http://%s/v1/companies/%s", addr, c.ID), bytes.NewReader(body))
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			resp, err = client.Do(ctx, http.MethodGet, fmt.Sprintf("http://%s/v1/companies/%s", addr, c.ID), bytes.NewReader(body))
			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	})
}
