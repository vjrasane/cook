package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func setupMockServer(t *testing.T) (*httptest.Server, *Client) {
	t.Helper()

	mux := http.NewServeMux()

	mux.HandleFunc("/api/loginextended", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", 405)
			return
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
			http.Error(w, "wrong content type", 400)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", 400)
			return
		}
		if r.Form.Get("grant_type") == "refresh_token" {
			if r.Form.Get("refresh_token") != "valid-refresh-token" {
				http.Error(w, "invalid refresh token", 401)
				return
			}
			json.NewEncoder(w).Encode(AuthResponse{
				AccessToken:  "refreshed-token-456",
				TokenType:    "Bearer",
				ExpiresIn:    86399,
				RefreshToken: "new-refresh-token",
			})
			return
		}
		if r.Form.Get("username") != "test@example.com" || r.Form.Get("password") != "secret" {
			http.Error(w, "invalid credentials", 401)
			return
		}
		json.NewEncoder(w).Encode(AuthResponse{
			AccessToken:  "test-token-123",
			TokenType:    "Bearer",
			ExpiresIn:    86399,
			RefreshToken: "valid-refresh-token",
		})
	})

	mux.HandleFunc("/api/lists", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token-123" {
			http.Error(w, "unauthorized", 401)
			return
		}
		switch r.Method {
		case "GET":
			json.NewEncoder(w).Encode([]List{
				{Id: "100", Name: "Groceries", Active: 1, Items: []Item{
					{Id: "1", IdAsNumber: 1, Name: "Milk", Checked: 0, Amount: "1", Unit: "L"},
				}},
				{Id: "200", Name: "Hardware", Active: 1},
			})
		case "POST":
			body, _ := io.ReadAll(r.Body)
			var req CreateListRequest
			json.Unmarshal(body, &req)
			w.WriteHeader(201)
			json.NewEncoder(w).Encode(List{
				Id:     "300",
				Name:   req.Name,
				Active: 1,
			})
		default:
			http.Error(w, "not found", 404)
		}
	})

	mux.HandleFunc("/api/lists/100", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token-123" {
			http.Error(w, "unauthorized", 401)
			return
		}
		switch r.Method {
		case "GET":
			json.NewEncoder(w).Encode(List{Id: "100", Name: "Groceries", Active: 1})
		case "DELETE":
			w.WriteHeader(200)
		case "PATCH":
			w.WriteHeader(200)
		default:
			http.Error(w, "not found", 404)
		}
	})

	mux.HandleFunc("/api/lists/100/items", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token-123" {
			http.Error(w, "unauthorized", 401)
			return
		}
		switch r.Method {
		case "GET":
			json.NewEncoder(w).Encode([]Item{
				{Id: "1", IdAsNumber: 1, Name: "Milk", Checked: 0, Amount: "1", Unit: "L"},
				{Id: "2", IdAsNumber: 2, Name: "Bread", Checked: 1},
			})
		case "POST":
			body, _ := io.ReadAll(r.Body)
			var req AddItemRequest
			json.Unmarshal(body, &req)
			w.WriteHeader(201)
			json.NewEncoder(w).Encode(Item{
				Id:         "3",
				IdAsNumber: 3,
				Name:       req.Name,
				Amount:     req.Amount,
				Unit:       req.Unit,
			})
		default:
			http.Error(w, "not found", 404)
		}
	})

	itemHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token-123" {
			http.Error(w, "unauthorized", 401)
			return
		}
		switch r.Method {
		case "PATCH":
			w.WriteHeader(200)
		case "DELETE":
			w.WriteHeader(200)
		default:
			http.Error(w, "not found", 404)
		}
	}
	mux.HandleFunc("/api/lists/100/items/1", itemHandler)
	mux.HandleFunc("/api/lists/100/items/2", itemHandler)

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	client := NewClient("test@example.com", "secret")
	client.BaseURL = server.URL

	return server, client
}

func TestAuthenticate(t *testing.T) {
	_, client := setupMockServer(t)

	if err := client.Authenticate(); err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if client.token != "test-token-123" {
		t.Errorf("token = %q, want %q", client.token, "test-token-123")
	}
}

func TestAuthenticateBadCredentials(t *testing.T) {
	_, client := setupMockServer(t)
	client.password = "wrong"

	if err := client.Authenticate(); err == nil {
		t.Fatal("Authenticate() expected error for bad credentials")
	}
}

func TestGetLists(t *testing.T) {
	_, client := setupMockServer(t)

	if err := client.Authenticate(); err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	lists, err := client.GetLists()
	if err != nil {
		t.Fatalf("GetLists() error = %v", err)
	}
	if len(lists) != 2 {
		t.Fatalf("GetLists() returned %d lists, want 2", len(lists))
	}
	if lists[0].Name != "Groceries" {
		t.Errorf("lists[0].Name = %q, want %q", lists[0].Name, "Groceries")
	}
	if lists[1].Id != "200" {
		t.Errorf("lists[1].Id = %q, want %q", lists[1].Id, "200")
	}
}

func TestGetList(t *testing.T) {
	_, client := setupMockServer(t)
	client.Authenticate()

	list, err := client.GetList("100")
	if err != nil {
		t.Fatalf("GetList() error = %v", err)
	}
	if list.Id != "100" {
		t.Errorf("list.Id = %q, want %q", list.Id, "100")
	}
	if list.Name != "Groceries" {
		t.Errorf("list.Name = %q, want %q", list.Name, "Groceries")
	}
}

func TestGetListItems(t *testing.T) {
	_, client := setupMockServer(t)
	client.Authenticate()

	items, err := client.GetListItems("100")
	if err != nil {
		t.Fatalf("GetListItems() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("GetListItems() returned %d items, want 2", len(items))
	}
	if items[0].Name != "Milk" {
		t.Errorf("items[0].Name = %q, want %q", items[0].Name, "Milk")
	}
	if items[1].Checked != 1 {
		t.Errorf("items[1].Checked = %d, want 1", items[1].Checked)
	}
}

func TestAddItem(t *testing.T) {
	_, client := setupMockServer(t)
	client.Authenticate()

	item, err := client.AddItem("100", AddItemRequest{
		Name:   "Eggs",
		Amount: "12",
		Unit:   "pcs",
	})
	if err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}
	if item.Name != "Eggs" {
		t.Errorf("item.Name = %q, want %q", item.Name, "Eggs")
	}
	if item.Amount != "12" {
		t.Errorf("item.Amount = %q, want %q", item.Amount, "12")
	}
}

func TestUpdateItem(t *testing.T) {
	_, client := setupMockServer(t)
	client.Authenticate()

	checked := 1
	err := client.UpdateItem("100", "1", UpdateItemRequest{Checked: &checked})
	if err != nil {
		t.Fatalf("UpdateItem() error = %v", err)
	}
}

func TestDeleteItem(t *testing.T) {
	_, client := setupMockServer(t)
	client.Authenticate()

	err := client.DeleteItem("100", "1")
	if err != nil {
		t.Fatalf("DeleteItem() error = %v", err)
	}
}

func TestResolveListIDNumeric(t *testing.T) {
	_, client := setupMockServer(t)
	client.Authenticate()

	id, err := client.ResolveListID("100")
	if err != nil {
		t.Fatalf("ResolveListID() error = %v", err)
	}
	if id != "100" {
		t.Errorf("ResolveListID() = %q, want %q", id, "100")
	}
}

func TestResolveListIDByName(t *testing.T) {
	_, client := setupMockServer(t)
	client.Authenticate()

	id, err := client.ResolveListID("Groceries")
	if err != nil {
		t.Fatalf("ResolveListID() error = %v", err)
	}
	if id != "100" {
		t.Errorf("ResolveListID() = %q, want %q", id, "100")
	}
}

func TestResolveListIDByNameCaseInsensitive(t *testing.T) {
	_, client := setupMockServer(t)
	client.Authenticate()

	id, err := client.ResolveListID("groceries")
	if err != nil {
		t.Fatalf("ResolveListID() error = %v", err)
	}
	if id != "100" {
		t.Errorf("ResolveListID() = %q, want %q", id, "100")
	}
}

func TestResolveListIDNotFound(t *testing.T) {
	_, client := setupMockServer(t)
	client.Authenticate()

	_, err := client.ResolveListID("NonExistent")
	if err == nil {
		t.Fatal("ResolveListID() expected error for non-existent list")
	}
}

func TestOutputSuccess(t *testing.T) {
	out := Output{Success: true, Data: map[string]string{"key": "value"}}
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var parsed map[string]any
	json.Unmarshal(b, &parsed)
	if parsed["success"] != true {
		t.Errorf("success = %v, want true", parsed["success"])
	}
}

func TestOutputError(t *testing.T) {
	out := Output{Success: false, Error: "something went wrong"}
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var parsed map[string]any
	json.Unmarshal(b, &parsed)
	if parsed["success"] != false {
		t.Errorf("success = %v, want false", parsed["success"])
	}
	if parsed["error"] != "something went wrong" {
		t.Errorf("error = %v, want %q", parsed["error"], "something went wrong")
	}
}

func TestUnauthenticatedRequest(t *testing.T) {
	_, client := setupMockServer(t)
	// Don't authenticate - token is empty

	_, err := client.GetLists()
	if err == nil {
		t.Fatal("GetLists() expected error for unauthenticated request")
	}
}

func TestAddItemMinimal(t *testing.T) {
	_, client := setupMockServer(t)
	client.Authenticate()

	item, err := client.AddItem("100", AddItemRequest{Name: "Butter"})
	if err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}
	if item.Name != "Butter" {
		t.Errorf("item.Name = %q, want %q", item.Name, "Butter")
	}
}

func TestCreateList(t *testing.T) {
	_, client := setupMockServer(t)
	client.Authenticate()

	list, err := client.CreateList("Party Supplies")
	if err != nil {
		t.Fatalf("CreateList() error = %v", err)
	}
	if list.Name != "Party Supplies" {
		t.Errorf("list.Name = %q, want %q", list.Name, "Party Supplies")
	}
	if list.Id != "300" {
		t.Errorf("list.Id = %q, want %q", list.Id, "300")
	}
}

func TestDeleteList(t *testing.T) {
	_, client := setupMockServer(t)
	client.Authenticate()

	err := client.DeleteList("100")
	if err != nil {
		t.Fatalf("DeleteList() error = %v", err)
	}
}

func TestUpdateList(t *testing.T) {
	_, client := setupMockServer(t)
	client.Authenticate()

	err := client.UpdateList("100", "Food")
	if err != nil {
		t.Fatalf("UpdateList() error = %v", err)
	}
}

func TestClearItemsAll(t *testing.T) {
	_, client := setupMockServer(t)
	client.Authenticate()

	deleted, err := client.ClearItems("100", false)
	if err != nil {
		t.Fatalf("ClearItems() error = %v", err)
	}
	if len(deleted) != 2 {
		t.Errorf("ClearItems() deleted %d items, want 2", len(deleted))
	}
}

func TestClearItemsCheckedOnly(t *testing.T) {
	_, client := setupMockServer(t)
	client.Authenticate()

	deleted, err := client.ClearItems("100", true)
	if err != nil {
		t.Fatalf("ClearItems() error = %v", err)
	}
	if len(deleted) != 1 {
		t.Errorf("ClearItems() deleted %d items, want 1", len(deleted))
	}
	if deleted[0] != "2" {
		t.Errorf("ClearItems() deleted item %q, want %q", deleted[0], "2")
	}
}

func TestServerError(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/loginextended" {
			json.NewEncoder(w).Encode(AuthResponse{AccessToken: "tok"})
			return
		}
		http.Error(w, "internal error", 500)
	}))
	defer server.Close()

	client := NewClient("a@b.com", "pass")
	client.BaseURL = server.URL
	client.Authenticate()

	_, err := client.GetLists()
	if err == nil {
		t.Fatal("expected error on server error")
	}
	fmt.Println("Got expected error:", err)
}

func TestRefreshToken(t *testing.T) {
	_, client := setupMockServer(t)

	err := client.refresh("valid-refresh-token")
	if err != nil {
		t.Fatalf("refresh() error = %v", err)
	}
	if client.token != "refreshed-token-456" {
		t.Errorf("token = %q, want %q", client.token, "refreshed-token-456")
	}
}

func TestRefreshTokenInvalid(t *testing.T) {
	_, client := setupMockServer(t)

	err := client.refresh("bad-token")
	if err == nil {
		t.Fatal("refresh() expected error for invalid refresh token")
	}
}

func TestAuthenticateUsesCachedToken(t *testing.T) {
	var loginCalls atomic.Int32

	mux := http.NewServeMux()
	mux.HandleFunc("/api/loginextended", func(w http.ResponseWriter, r *http.Request) {
		loginCalls.Add(1)
		json.NewEncoder(w).Encode(AuthResponse{
			AccessToken: "server-token",
			ExpiresIn:   86399,
		})
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)

	saveTokenCache(&TokenCache{
		AccessToken:  "cached-token-abc",
		RefreshToken: "some-refresh",
		ExpiresAt:    time.Now().Unix() + 3600,
	})

	client := NewClient("test@example.com", "secret")
	client.BaseURL = server.URL

	if err := client.Authenticate(); err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if client.token != "cached-token-abc" {
		t.Errorf("token = %q, want %q", client.token, "cached-token-abc")
	}
	if loginCalls.Load() != 0 {
		t.Errorf("login endpoint called %d times, want 0", loginCalls.Load())
	}
}

func TestAuthenticateRefreshesExpiredToken(t *testing.T) {
	_, client := setupMockServer(t)

	saveTokenCache(&TokenCache{
		AccessToken:  "expired-token",
		RefreshToken: "valid-refresh-token",
		ExpiresAt:    time.Now().Unix() - 1,
	})

	if err := client.Authenticate(); err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if client.token != "refreshed-token-456" {
		t.Errorf("token = %q, want %q", client.token, "refreshed-token-456")
	}
}

func TestAuthenticateFallsBackToLogin(t *testing.T) {
	_, client := setupMockServer(t)

	saveTokenCache(&TokenCache{
		AccessToken:  "expired-token",
		RefreshToken: "bad-refresh-token",
		ExpiresAt:    time.Now().Unix() - 1,
	})

	if err := client.Authenticate(); err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if client.token != "test-token-123" {
		t.Errorf("token = %q, want %q", client.token, "test-token-123")
	}
}
