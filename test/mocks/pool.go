package mocks

import (
	"context"
	"database/sql"

	"github.com/kgjoner/sphinx/internal/common"
)

// MockBasePool mocks the repository factory interface
type MockBasePool struct {
	mockQueries *MockQueries
}

func NewMockBasePool() *MockBasePool {
	return &MockBasePool{
		mockQueries: NewMockQueries(),
	}
}

func (m *MockBasePool) NewDAO(ctx context.Context) common.BaseRepo {
	return m.mockQueries
}

func (m *MockBasePool) GetMockQueries() *MockQueries {
	return m.mockQueries
}

func (m *MockBasePool) Close() error {
	m.mockQueries.Clear()
	return nil
}

func (m *MockBasePool) WithTransaction(ctx context.Context, opts *sql.TxOptions, fn func(common.BaseRepo) (any, error)) (any, error) {
	return fn(m.mockQueries)
}

func (m *MockBasePool) WithReadOnlyTransaction(ctx context.Context, fn func(common.BaseRepo) (any, error)) (any, error) {
	return fn(m.mockQueries)
}
