package jsonapi

// Href is a URL.
type Href struct {
	Href string `json:"href,omitempty"`
}

// PageLinks is the HAL links for a paginated result.
type PageLinks struct {
	Self *Href `json:"self,omitempty"`
	Page *Href `json:"page,omitempty"`
}

// ResourceLinks is the HAL links for a single result.
type ResourceLinks struct {
	Self  *Href `json:"self,omitempty"`
	Steps *Href `json:"steps,omitempty"`
}

// NewHref returns a new Href.
func NewHref(uri string) *Href {
	if uri == "" {
		return nil
	}

	return &Href{Href: uri}
}
