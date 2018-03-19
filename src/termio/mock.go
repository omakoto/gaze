package termio

import (
	"github.com/stretchr/testify/mock"
	"time"
)

type MockTerm struct {
	mock.Mock
}

func (m *MockTerm) Width() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTerm) Height() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTerm) Clear() {
	m.Called()
}

func (m *MockTerm) Flush() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTerm) Finish() {
	m.Called()
}

func (m *MockTerm) WriteZeroWidthString(s string) {
	m.Called(s)
}

func (m *MockTerm) WriteZeroWidthBytes(bytes []byte) {
	m.Called(bytes)
}

func (m *MockTerm) MoveTo(newX, newY int) {
	m.Called(newX, newY)
}

func (m *MockTerm) Tab() {
	m.Called()
}

func (m *MockTerm) CanWrite() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTerm) CanWriteChars(charWidth int) bool {
	args := m.Called(charWidth)
	return args.Bool(0)
}

func (m *MockTerm) NewLine() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTerm) WriteString(s string) bool {
	args := m.Called(s)
	return args.Bool(0)
}

func (m *MockTerm) WriteRune(ch rune) bool {
	args := m.Called(ch)
	return args.Bool(0)
}

func (m *MockTerm) ReadByteTimeout(timeout time.Duration) (byte, error) {
	args := m.Called(timeout)
	return args.Get(0).(byte), args.Error(1)
}
