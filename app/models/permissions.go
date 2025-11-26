package models

type Permission struct {
	ID          string
	Name        string
	Resource    string
	Action      string
	Description string
}
