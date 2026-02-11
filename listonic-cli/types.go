package main

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type List struct {
	Id      string `json:"Id"`
	Name    string `json:"Name"`
	Active  int    `json:"Active"`
	Deleted int    `json:"Deleted"`
	Items   []Item `json:"Items"`
}

type Item struct {
	Id         string  `json:"Id"`
	IdAsNumber int64   `json:"IdAsNumber"`
	Name       string  `json:"Name"`
	Checked    int     `json:"Checked"`
	Amount     string  `json:"Amount"`
	Unit       string  `json:"Unit"`
	Price      float64 `json:"Price"`
	Description string `json:"Description"`
	CategoryId  int    `json:"CategoryId"`
}

type AddItemRequest struct {
	Name   string `json:"Name"`
	Amount string `json:"Amount,omitempty"`
	Unit   string `json:"Unit,omitempty"`
}

type UpdateItemRequest struct {
	Checked *int `json:"Checked,omitempty"`
}

type CreateListRequest struct {
	Name string `json:"Name"`
}

type UpdateListRequest struct {
	Name string `json:"Name"`
}
