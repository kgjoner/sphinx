package mocks

import (
	"context"
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

func (m *MockBasePool) NewQueries(ctx context.Context) common.BaseRepo {
	return m.mockQueries
}

func (m *MockBasePool) GetMockQueries() *MockQueries {
	return m.mockQueries
}
