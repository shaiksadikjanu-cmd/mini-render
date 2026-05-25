package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

var (
	SupabaseURL    string
	SupabaseSecret string
)

func Init() {
	SupabaseURL = os.Getenv("SUPABASE_URL")
	SupabaseSecret = os.Getenv("SUPABASE_SECRET_KEY")
}

func request(method, path string, body interface{}) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		reqBody = bytes.NewReader(b)
	}

	url := SupabaseURL + "/rest/v1/" + path
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("apikey", SupabaseSecret)
	req.Header.Set("Authorization", "Bearer "+SupabaseSecret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	return data, resp.StatusCode, err
}

// Insert a row and return the created row
func Insert(table string, row interface{}) ([]byte, error) {
	data, status, err := request("POST", table, row)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("supabase error %d: %s", status, string(data))
	}
	return data, nil
}

// Select rows with a filter
func Select(table, filter string) ([]byte, error) {
	path := table
	if filter != "" {
		path += "?" + filter
	}
	data, status, err := request("GET", path, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("supabase error %d: %s", status, string(data))
	}
	return data, nil
}

// Update rows with a filter
func Update(table, filter string, updates interface{}) ([]byte, error) {
	path := table + "?" + filter
	data, status, err := request("PATCH", path, updates)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("supabase error %d: %s", status, string(data))
	}
	return data, nil
}
