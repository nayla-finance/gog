package pagination

import (
	"testing"
)

func TestToPaginatedResponse(t *testing.T) {
	tests := []struct {
		name         string
		data         []string
		totalRecords int
		page         int
		pageSize     int
		wantStatus   string
		wantHasNext  bool
		wantHasPrev  bool
		wantPages    int
	}{
		{
			name:         "First page with more pages",
			data:         []string{"a", "b", "c"},
			totalRecords: 10,
			page:         1,
			pageSize:     3,
			wantStatus:   "success",
			wantHasNext:  true,
			wantHasPrev:  false,
			wantPages:    4,
		},
		{
			name:         "Middle page",
			data:         []string{"d", "e", "f"},
			totalRecords: 10,
			page:         2,
			pageSize:     3,
			wantStatus:   "success",
			wantHasNext:  true,
			wantHasPrev:  true,
			wantPages:    4,
		},
		{
			name:         "Last page",
			data:         []string{"j"},
			totalRecords: 10,
			page:         4,
			pageSize:     3,
			wantStatus:   "success",
			wantHasNext:  false,
			wantHasPrev:  true,
			wantPages:    4,
		},
		{
			name:         "Empty data",
			data:         []string{},
			totalRecords: 0,
			page:         1,
			pageSize:     10,
			wantStatus:   "success",
			wantHasNext:  false,
			wantHasPrev:  false,
			wantPages:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToPaginatedResponse(tt.data, tt.totalRecords, tt.page, tt.pageSize)

			if got.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", got.Status, tt.wantStatus)
			}

			if len(got.Data) != len(tt.data) {
				t.Errorf("Data length = %v, want %v", len(got.Data), len(tt.data))
			}

			if got.Pagination.HasNext != tt.wantHasNext {
				t.Errorf("HasNext = %v, want %v", got.Pagination.HasNext, tt.wantHasNext)
			}

			if got.Pagination.HasPrevious != tt.wantHasPrev {
				t.Errorf("HasPrevious = %v, want %v", got.Pagination.HasPrevious, tt.wantHasPrev)
			}

			if got.Pagination.TotalPages != tt.wantPages {
				t.Errorf("TotalPages = %v, want %v", got.Pagination.TotalPages, tt.wantPages)
			}

			if got.Pagination.Page != tt.page {
				t.Errorf("Page = %v, want %v", got.Pagination.Page, tt.page)
			}

			if got.Pagination.PageSize != tt.pageSize {
				t.Errorf("PageSize = %v, want %v", got.Pagination.PageSize, tt.pageSize)
			}

			if got.Pagination.TotalRecords != tt.totalRecords {
				t.Errorf("TotalRecords = %v, want %v", got.Pagination.TotalRecords, tt.totalRecords)
			}
		})
	}
}
func TestGetPaginationParams(t *testing.T) {
	tests := []struct {
		name      string
		page      int
		limit     int
		wantPage  int
		wantLimit int
	}{
		{
			name:      "Valid page and limit",
			page:      2,
			limit:     10,
			wantPage:  2,
			wantLimit: 10,
		},
		{
			name:      "Zero page and limit",
			page:      0,
			limit:     0,
			wantPage:  1,
			wantLimit: 25,
		},
		{
			name:      "Negative page and limit",
			page:      -1,
			limit:     -25,
			wantPage:  1,
			wantLimit: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPage, gotLimit := GetPaginationParams(tt.page, tt.limit)

			if gotPage != tt.wantPage {
				t.Errorf("Page = %v, want %v", gotPage, tt.wantPage)
			}

			if gotLimit != tt.wantLimit {
				t.Errorf("Limit = %v, want %v", gotLimit, tt.wantLimit)
			}
		})
	}
}
