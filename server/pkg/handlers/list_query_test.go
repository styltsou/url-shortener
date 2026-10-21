package handlers

import (
	"net/http/httptest"
	"testing"
)

func TestListLinksQueryBounds(t *testing.T) {
	for _, rawQuery := range []string{"limit=101", "page=2147483647", "page=-1", "limit=0"} {
		req := httptest.NewRequest("GET", "/api/v1/links?"+rawQuery, nil)
		_, err := parseListLinksQuery(req)
		if err == nil {
			t.Fatalf("%s should be rejected", rawQuery)
		}
	}
	req := httptest.NewRequest("GET", "/api/v1/links?page=21474836&limit=100", nil)
	query, err := parseListLinksQuery(req)
	if err != nil || query.Page != 21474836 || query.Limit != 100 {
		t.Fatalf("valid boundary rejected: query=%+v err=%+v", query, err)
	}
}
