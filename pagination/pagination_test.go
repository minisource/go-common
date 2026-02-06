package pagination

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestParseParams(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		params := ParseParams(c)
		assert.Equal(t, 1, params.Page)
		assert.Equal(t, 20, params.PerPage)
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	app.Test(req)
}

func TestParseParamsWithValues(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		params := ParseParams(c)
		assert.Equal(t, 2, params.Page)
		assert.Equal(t, 50, params.PerPage)
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test?page=2&per_page=50", nil)
	app.Test(req)
}

func TestParseParamsMaxPerPage(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		params := ParseParams(c)
		assert.Equal(t, MaxPageSize, params.PerPage)
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test?per_page=1000", nil)
	app.Test(req)
}

func TestParseParamsInvalidPage(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		params := ParseParams(c)
		assert.Equal(t, 1, params.Page)
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test?page=-1", nil)
	app.Test(req)
}

func TestParamsOffset(t *testing.T) {
	testCases := []struct {
		page    int
		perPage int
		offset  int
	}{
		{1, 20, 0},
		{2, 20, 20},
		{3, 50, 100},
		{1, 10, 0},
	}

	for _, tc := range testCases {
		params := Params{Page: tc.page, PerPage: tc.perPage}
		assert.Equal(t, tc.offset, params.Offset())
	}
}

func TestNewResult(t *testing.T) {
	result := NewResult(1, 20, 100)

	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 20, result.PerPage)
	assert.Equal(t, int64(100), result.Total)
	assert.Equal(t, 5, result.TotalPages)
	assert.True(t, result.HasNext)
	assert.False(t, result.HasPrev)
}

func TestNewResultLastPage(t *testing.T) {
	result := NewResult(5, 20, 100)

	assert.Equal(t, 5, result.Page)
	assert.False(t, result.HasNext)
	assert.True(t, result.HasPrev)
}

func TestCursorEncoding(t *testing.T) {
	cursorData := CursorData{
		ID:        "123",
		CreatedAt: 1234567890,
		Value:     "test",
	}

	encoded := EncodeCursor(cursorData)
	assert.NotEmpty(t, encoded)

	decoded, err := DecodeCursor(encoded)
	assert.NoError(t, err)
	assert.Equal(t, cursorData.ID, decoded.ID)
	assert.Equal(t, cursorData.CreatedAt, decoded.CreatedAt)
	assert.Equal(t, cursorData.Value, decoded.Value)
}

func TestDecodeCursorEmpty(t *testing.T) {
	decoded, err := DecodeCursor("")
	assert.NoError(t, err)
	assert.Nil(t, decoded)
}
