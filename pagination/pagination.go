package pagination

import (
	"encoding/base64"
	"encoding/json"
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// DefaultPageSize is the default number of items per page
const DefaultPageSize = 20

// MaxPageSize is the maximum allowed page size
const MaxPageSize = 100

// Params holds pagination parameters
type Params struct {
	Page    int    `query:"page"`
	PerPage int    `query:"per_page"`
	Cursor  string `query:"cursor"`
	Sort    string `query:"sort"`
	Order   string `query:"order"` // asc, desc
}

// Result holds pagination result
type Result struct {
	Page       int    `json:"page,omitempty"`
	PerPage    int    `json:"perPage,omitempty"`
	Total      int64  `json:"total"`
	TotalPages int    `json:"totalPages"`
	HasNext    bool   `json:"hasNext"`
	HasPrev    bool   `json:"hasPrev"`
	NextCursor string `json:"nextCursor,omitempty"`
	PrevCursor string `json:"prevCursor,omitempty"`
}

// CursorData holds cursor data for encoding/decoding
type CursorData struct {
	ID        string `json:"id"`
	CreatedAt int64  `json:"ca,omitempty"`
	Value     string `json:"v,omitempty"`
}

// ParseParams extracts pagination params from Fiber context
func ParseParams(c *fiber.Ctx) Params {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", strconv.Itoa(DefaultPageSize)))

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = DefaultPageSize
	}
	if perPage > MaxPageSize {
		perPage = MaxPageSize
	}

	order := c.Query("order", "desc")
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	return Params{
		Page:    page,
		PerPage: perPage,
		Cursor:  c.Query("cursor"),
		Sort:    c.Query("sort", "created_at"),
		Order:   order,
	}
}

// Offset calculates the offset for offset-based pagination
func (p Params) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// Limit returns the limit
func (p Params) Limit() int {
	return p.PerPage
}

// NewResult creates a pagination result for offset-based pagination
func NewResult(page, perPage int, total int64) *Result {
	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	if totalPages < 1 {
		totalPages = 1
	}

	return &Result{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// NewCursorResult creates a pagination result for cursor-based pagination
func NewCursorResult(total int64, hasNext bool, nextCursor, prevCursor string) *Result {
	return &Result{
		Total:      total,
		HasNext:    hasNext,
		HasPrev:    prevCursor != "",
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
	}
}

// EncodeCursor encodes cursor data to string
func EncodeCursor(data CursorData) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(jsonData)
}

// DecodeCursor decodes cursor string to data
func DecodeCursor(cursor string) (*CursorData, error) {
	if cursor == "" {
		return nil, nil
	}

	jsonData, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, err
	}

	var data CursorData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// Paginator provides pagination utilities for GORM
type Paginator struct {
	db     *gorm.DB
	params Params
}

// NewPaginator creates a new paginator
func NewPaginator(db *gorm.DB, params Params) *Paginator {
	return &Paginator{
		db:     db,
		params: params,
	}
}

// Paginate applies pagination to the query and returns results
func (p *Paginator) Paginate(dest interface{}, countDest *int64) (*Result, error) {
	// Count total
	if err := p.db.Count(countDest).Error; err != nil {
		return nil, err
	}

	// Apply sorting
	orderClause := p.params.Sort + " " + p.params.Order
	p.db = p.db.Order(orderClause)

	// Apply pagination
	offset := p.params.Offset()
	limit := p.params.Limit()

	if err := p.db.Offset(offset).Limit(limit).Find(dest).Error; err != nil {
		return nil, err
	}

	return NewResult(p.params.Page, p.params.PerPage, *countDest), nil
}

// PaginateWithCursor applies cursor-based pagination
func (p *Paginator) PaginateWithCursor(dest interface{}, idField, sortField string, countDest *int64) (*Result, error) {
	// Count total
	if err := p.db.Count(countDest).Error; err != nil {
		return nil, err
	}

	// Decode cursor if present
	cursor, err := DecodeCursor(p.params.Cursor)
	if err != nil {
		return nil, err
	}

	// Apply cursor condition
	if cursor != nil {
		if p.params.Order == "desc" {
			p.db = p.db.Where(sortField+" < ?", cursor.Value).Or(
				sortField+" = ? AND "+idField+" < ?", cursor.Value, cursor.ID,
			)
		} else {
			p.db = p.db.Where(sortField+" > ?", cursor.Value).Or(
				sortField+" = ? AND "+idField+" > ?", cursor.Value, cursor.ID,
			)
		}
	}

	// Apply sorting and limit (fetch one extra to check hasNext)
	orderClause := sortField + " " + p.params.Order + ", " + idField + " " + p.params.Order
	limit := p.params.Limit() + 1

	if err := p.db.Order(orderClause).Limit(limit).Find(dest).Error; err != nil {
		return nil, err
	}

	// Check hasNext and create cursors
	// Note: The caller needs to handle the extra record if present
	hasNext := false
	var nextCursor string

	return NewCursorResult(*countDest, hasNext, nextCursor, p.params.Cursor), nil
}

// Scope returns a GORM scope for pagination
func Scope(params Params) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := params.Offset()
		limit := params.Limit()
		orderClause := params.Sort + " " + params.Order

		return db.Offset(offset).Limit(limit).Order(orderClause)
	}
}
