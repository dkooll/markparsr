package markparsr

import (
	"errors"
	"testing"
)

func TestErrorCollector_Add(t *testing.T) {
	tests := []struct {
		name          string
		errorsToAdd   []error
		expectedCount int
	}{
		{
			name:          "add single error",
			errorsToAdd:   []error{errors.New("test error")},
			expectedCount: 1,
		},
		{
			name:          "add nil error",
			errorsToAdd:   []error{nil},
			expectedCount: 0,
		},
		{
			name: "add multiple errors",
			errorsToAdd: []error{
				errors.New("error 1"),
				errors.New("error 2"),
				errors.New("error 3"),
			},
			expectedCount: 3,
		},
		{
			name: "add mixed nil and valid errors",
			errorsToAdd: []error{
				errors.New("error 1"),
				nil,
				errors.New("error 2"),
				nil,
			},
			expectedCount: 2,
		},
		{
			name:          "no errors added",
			errorsToAdd:   []error{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := &ErrorCollector{}
			for _, err := range tt.errorsToAdd {
				collector.Add(err)
			}
			if len(collector.Errors()) != tt.expectedCount {
				t.Errorf("ErrorCollector.Add() resulted in %d errors; want %d", len(collector.Errors()), tt.expectedCount)
			}
		})
	}
}

func TestErrorCollector_AddMany(t *testing.T) {
	tests := []struct {
		name          string
		errorSets     [][]error
		expectedCount int
	}{
		{
			name: "add multiple error sets",
			errorSets: [][]error{
				{errors.New("error 1"), errors.New("error 2")},
				{errors.New("error 3")},
			},
			expectedCount: 3,
		},
		{
			name: "add empty error set",
			errorSets: [][]error{
				{},
			},
			expectedCount: 0,
		},
		{
			name: "add error sets with nils",
			errorSets: [][]error{
				{errors.New("error 1"), nil, errors.New("error 2")},
				{nil, errors.New("error 3")},
			},
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := &ErrorCollector{}
			for _, errorSet := range tt.errorSets {
				collector.AddMany(errorSet)
			}
			if len(collector.Errors()) != tt.expectedCount {
				t.Errorf("ErrorCollector.AddMany() resulted in %d errors; want %d", len(collector.Errors()), tt.expectedCount)
			}
		})
	}
}

func TestErrorCollector_Errors(t *testing.T) {
	t.Run("returns all collected errors", func(t *testing.T) {
		collector := &ErrorCollector{}
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")

		collector.Add(err1)
		collector.Add(err2)

		errs := collector.Errors()
		if len(errs) != 2 {
			t.Errorf("ErrorCollector.Errors() returned %d errors; want 2", len(errs))
		}
		if errs[0] != err1 {
			t.Errorf("ErrorCollector.Errors()[0] = %v; want %v", errs[0], err1)
		}
		if errs[1] != err2 {
			t.Errorf("ErrorCollector.Errors()[1] = %v; want %v", errs[1], err2)
		}
	})

	t.Run("returns empty slice when no errors", func(t *testing.T) {
		collector := &ErrorCollector{}
		errs := collector.Errors()
		if len(errs) != 0 {
			t.Errorf("ErrorCollector.Errors() returned %d errors; want 0", len(errs))
		}
	})
}

func TestErrorCollector_HasErrors(t *testing.T) {
	tests := []struct {
		name        string
		errorsToAdd []error
		expectedHas bool
	}{
		{
			name:        "has errors when errors added",
			errorsToAdd: []error{errors.New("test error")},
			expectedHas: true,
		},
		{
			name:        "no errors when empty",
			errorsToAdd: []error{},
			expectedHas: false,
		},
		{
			name:        "no errors when only nil added",
			errorsToAdd: []error{nil, nil},
			expectedHas: false,
		},
		{
			name: "has errors with mixed nil and valid",
			errorsToAdd: []error{
				nil,
				errors.New("error"),
				nil,
			},
			expectedHas: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := &ErrorCollector{}
			for _, err := range tt.errorsToAdd {
				collector.Add(err)
			}
			if collector.HasErrors() != tt.expectedHas {
				t.Errorf("ErrorCollector.HasErrors() = %v; want %v", collector.HasErrors(), tt.expectedHas)
			}
		})
	}
}
