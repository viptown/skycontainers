package pagination

import (
	"math"
)

type Pager struct {
	TotalItems  int
	CurrentPage int
	PageSize    int
	TotalPages  int
	StartPage   int
	EndPage     int
	Pages       []int
}

func NewPager(totalItems, currentPage, pageSize int) Pager {
	totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))
	if currentPage < 1 {
		currentPage = 1
	}
	if totalPages > 0 && currentPage > totalPages {
		currentPage = totalPages
	}

	startPage := currentPage - 2
	endPage := currentPage + 2

	if startPage <= 0 {
		endPage -= (startPage - 1)
		startPage = 1
	}

	if endPage > totalPages {
		endPage = totalPages
		if endPage > 5 {
			startPage = endPage - 4
		}
	}

	var pages []int
	for i := startPage; i <= endPage; i++ {
		pages = append(pages, i)
	}

	return Pager{
		TotalItems:  totalItems,
		CurrentPage: currentPage,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		StartPage:   startPage,
		EndPage:     endPage,
		Pages:       pages,
	}
}

func (p Pager) Offset() int {
	if p.CurrentPage < 1 {
		return 0
	}
	return (p.CurrentPage - 1) * p.PageSize
}
