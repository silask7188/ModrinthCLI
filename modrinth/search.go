package modrinth

import (
	"fmt"
	"net/url"
)
type SearchResponse struct {
	Hits      []Project `json:"hits"`       // how many results
	Offset    int       `json:"offset"`   // pagination
	Limit     int       `json:"limit"` // pagination
	TotalHits int       `json:"total_hits"` // pagination
}

type SearchParams struct {
	Query  string   // "sodium"
	Facets []string // ["categories=fabric", "project_type=mod"]
	Offset int      // pagination
	Limit  int      //pagination
}


func (p SearchParams) Values() url.Values {
	v := make(url.Values)

		if p.Query != "" {
		v.Set("query", p.Query)
	}
	for _, f := range p.Facets {
		v.Add("facets", fmt.Sprintf(`["%s"]`, f))
	}
	if p.Offset > 0 {
		v.Set("offset", fmt.Sprint(p.Offset))
	}
	if p.Limit > 0 {
		v.Set("limit", fmt.Sprint(p.Limit))
	}
	return v
}
