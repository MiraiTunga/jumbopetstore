package models

type Category struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type Tag struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type Pet struct {
	Id        int64    `json:"id"`
	Category  Category `json:"category,omitempty"`
	Name      string   `json:"name,omitempty"`
	PhotoUrls []string `json:"photoUrls,omitempty"`
	Tags      *[]Tag    `json:"tags,omitempty"`
	Status    string   `json:"status,omitempty"`
}
