package jsonapi

// DataResponse is a common struct for single response.
type DataResponse struct {
	Data  interface{}   `json:"data"`
	Links ResourceLinks `json:"links,omitempty"`
}

// LinksResponse is a common struct for embedding links response.
type LinksResponse struct {
	Links ResourceLinks `json:"links,omitempty"`
}

// PageResponse is a common struct for paged response.
type PageResponse struct {
	Data  []interface{} `json:"data"`
	Links PageLinks     `json:"links,omitempty"`
}

// ErrorResponse is a common struct for error response.
type ErrorResponse struct {
	Errors []string `json:"errors"`
}
