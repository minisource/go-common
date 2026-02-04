package testing

import (
	"time"

	"github.com/google/uuid"
)

// ============================================
// Fixture Generators
// ============================================

// FixtureBuilder helps build test data
type FixtureBuilder struct {
	counter int
}

// NewFixtureBuilder creates a new fixture builder
func NewFixtureBuilder() *FixtureBuilder {
	return &FixtureBuilder{}
}

// NextID returns a new UUID
func (f *FixtureBuilder) NextID() uuid.UUID {
	return uuid.New()
}

// NextInt returns an incrementing integer
func (f *FixtureBuilder) NextInt() int {
	f.counter++
	return f.counter
}

// NextEmail generates a unique email
func (f *FixtureBuilder) NextEmail() string {
	f.counter++
	return "test" + string(rune('0'+f.counter)) + "@example.com"
}

// NextPhone generates a unique phone number
func (f *FixtureBuilder) NextPhone() string {
	f.counter++
	return "+1555000" + padLeft(f.counter, 4)
}

// padLeft pads integer with zeros
func padLeft(n, width int) string {
	s := ""
	for i := 0; i < width; i++ {
		s = string(rune('0'+(n%10))) + s
		n /= 10
	}
	return s
}

// ============================================
// Common Test Data
// ============================================

// TestUser represents a test user
type TestUser struct {
	ID        uuid.UUID
	Email     string
	Name      string
	Password  string
	Role      string
	TenantID  uuid.UUID
	CreatedAt time.Time
}

// DefaultTestUser returns a default test user
func DefaultTestUser() TestUser {
	return TestUser{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Name:      "Test User",
		Password:  "password123",
		Role:      "user",
		TenantID:  uuid.New(),
		CreatedAt: time.Now(),
	}
}

// TestUserBuilder builds test users
type TestUserBuilder struct {
	user TestUser
}

// NewTestUserBuilder creates a new test user builder
func NewTestUserBuilder() *TestUserBuilder {
	return &TestUserBuilder{
		user: DefaultTestUser(),
	}
}

// WithID sets the ID
func (b *TestUserBuilder) WithID(id uuid.UUID) *TestUserBuilder {
	b.user.ID = id
	return b
}

// WithEmail sets the email
func (b *TestUserBuilder) WithEmail(email string) *TestUserBuilder {
	b.user.Email = email
	return b
}

// WithName sets the name
func (b *TestUserBuilder) WithName(name string) *TestUserBuilder {
	b.user.Name = name
	return b
}

// WithRole sets the role
func (b *TestUserBuilder) WithRole(role string) *TestUserBuilder {
	b.user.Role = role
	return b
}

// WithTenant sets the tenant ID
func (b *TestUserBuilder) WithTenant(tenantID uuid.UUID) *TestUserBuilder {
	b.user.TenantID = tenantID
	return b
}

// Build returns the test user
func (b *TestUserBuilder) Build() TestUser {
	return b.user
}

// ============================================
// Time Helpers
// ============================================

// FixedTime returns a fixed time for testing
func FixedTime() time.Time {
	return time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
}

// TimeAfter returns time after duration
func TimeAfter(d time.Duration) time.Time {
	return FixedTime().Add(d)
}

// TimeBefore returns time before duration
func TimeBefore(d time.Duration) time.Time {
	return FixedTime().Add(-d)
}

// ============================================
// Slice Helpers
// ============================================

// GenerateIDs generates n UUIDs
func GenerateIDs(n int) []uuid.UUID {
	ids := make([]uuid.UUID, n)
	for i := 0; i < n; i++ {
		ids[i] = uuid.New()
	}
	return ids
}

// GenerateStrings generates n strings with prefix
func GenerateStrings(prefix string, n int) []string {
	strs := make([]string, n)
	for i := 0; i < n; i++ {
		strs[i] = prefix + padLeft(i+1, 3)
	}
	return strs
}
