package markparsr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewFileValidator(t *testing.T) {
	tests := []struct {
		name            string
		readmePath      string
		modulePath      string
		additionalFiles []string
		expectedReq     int
		expectedAdd     int
	}{
		{
			name:            "basic validator with no additional files",
			readmePath:      "/tmp/README.md",
			modulePath:      "/tmp",
			additionalFiles: []string{},
			expectedReq:     4, // README.md, outputs.tf, variables.tf, terraform.tf
			expectedAdd:     0,
		},
		{
			name:            "validator with additional files",
			readmePath:      "/tmp/README.md",
			modulePath:      "/tmp",
			additionalFiles: []string{"main.tf", "versions.tf"},
			expectedReq:     4,
			expectedAdd:     2,
		},
		{
			name:            "validator with absolute additional paths",
			readmePath:      "/tmp/README.md",
			modulePath:      "/tmp",
			additionalFiles: []string{"/tmp/main.tf"},
			expectedReq:     4,
			expectedAdd:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fv := NewFileValidator(tt.readmePath, tt.modulePath, tt.additionalFiles)

			if fv == nil {
				t.Fatal("NewFileValidator() returned nil")
			}

			if len(fv.requiredFiles) != tt.expectedReq {
				t.Errorf("NewFileValidator() required files = %d; want %d", len(fv.requiredFiles), tt.expectedReq)
			}

			if len(fv.additionalFiles) != tt.expectedAdd {
				t.Errorf("NewFileValidator() additional files = %d; want %d", len(fv.additionalFiles), tt.expectedAdd)
			}

			if fv.rootDir != tt.modulePath {
				t.Errorf("NewFileValidator() rootDir = %q; want %q", fv.rootDir, tt.modulePath)
			}
		})
	}
}

func TestFileValidator_Validate(t *testing.T) {
	tmpDir := t.TempDir()

	readmePath := filepath.Join(tmpDir, "README.md")
	outputsPath := filepath.Join(tmpDir, "outputs.tf")
	variablesPath := filepath.Join(tmpDir, "variables.tf")
	terraformPath := filepath.Join(tmpDir, "terraform.tf")
	mainPath := filepath.Join(tmpDir, "main.tf")

	os.WriteFile(readmePath, []byte("# Test Module"), 0o644)
	os.WriteFile(outputsPath, []byte("output \"test\" { value = \"test\" }"), 0o644)
	os.WriteFile(variablesPath, []byte("variable \"test\" { type = string }"), 0o644)
	os.WriteFile(terraformPath, []byte("terraform { required_version = \">= 1.0\" }"), 0o644)
	os.WriteFile(mainPath, []byte("resource \"test\" \"test\" {}"), 0o644)

	tests := []struct {
		name            string
		setupFiles      func()
		additionalFiles []string
		expectedErrors  int
		errorContains   []string
	}{
		{
			name: "all required files present",
			setupFiles: func() {
				// Files already created
			},
			additionalFiles: []string{},
			expectedErrors:  0,
		},
		{
			name: "missing required file",
			setupFiles: func() {
				os.Remove(outputsPath)
			},
			additionalFiles: []string{},
			expectedErrors:  1,
			errorContains:   []string{"outputs.tf", "does not exist"},
		},
		{
			name: "empty required file",
			setupFiles: func() {
				os.WriteFile(outputsPath, []byte{}, 0o644)
			},
			additionalFiles: []string{},
			expectedErrors:  1,
			errorContains:   []string{"outputs.tf", "empty"},
		},
		{
			name: "missing additional file",
			setupFiles: func() {
				os.Remove(filepath.Join(tmpDir, "extra.tf"))
			},
			additionalFiles: []string{"extra.tf"},
			expectedErrors:  1,
			errorContains:   []string{"extra.tf", "does not exist"},
		},
		{
			name: "all files including additional present",
			setupFiles: func() {
				os.WriteFile(outputsPath, []byte("output \"test\" { value = \"test\" }"), 0o644)
			},
			additionalFiles: []string{"main.tf"},
			expectedErrors:  0,
		},
		{
			name: "multiple missing files",
			setupFiles: func() {
				os.Remove(variablesPath)
				os.Remove(terraformPath)
			},
			additionalFiles: []string{},
			expectedErrors:  2,
			errorContains:   []string{"variables.tf", "terraform.tf"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.WriteFile(readmePath, []byte("# Test Module"), 0o644)
			os.WriteFile(outputsPath, []byte("output \"test\" { value = \"test\" }"), 0o644)
			os.WriteFile(variablesPath, []byte("variable \"test\" { type = string }"), 0o644)
			os.WriteFile(terraformPath, []byte("terraform { required_version = \">= 1.0\" }"), 0o644)
			os.WriteFile(mainPath, []byte("resource \"test\" \"test\" {}"), 0o644)

			tt.setupFiles()

			fv := NewFileValidator(readmePath, tmpDir, tt.additionalFiles)
			errs := fv.Validate()

			if len(errs) != tt.expectedErrors {
				t.Errorf("Validate() returned %d errors; want %d", len(errs), tt.expectedErrors)
				for i, err := range errs {
					t.Logf("  error %d: %v", i+1, err)
				}
			}

			for _, substr := range tt.errorContains {
				found := false
				for _, err := range errs {
					if err != nil && strings.Contains(err.Error(), substr) {
						found = true
						break
					}
				}
				if !found && tt.expectedErrors > 0 {
					t.Errorf("Expected error containing %q, but not found", substr)
				}
			}
		})
	}
}

func TestValidateFile(t *testing.T) {
	tmpDir := t.TempDir()

	existingFile := filepath.Join(tmpDir, "existing.tf")
	emptyFile := filepath.Join(tmpDir, "empty.tf")
	nonExistentFile := filepath.Join(tmpDir, "nonexistent.tf")

	os.WriteFile(existingFile, []byte("content"), 0o644)
	os.WriteFile(emptyFile, []byte{}, 0o644)

	tests := []struct {
		name          string
		filePath      string
		expectError   bool
		errorContains string
	}{
		{
			name:        "existing file with content",
			filePath:    existingFile,
			expectError: false,
		},
		{
			name:          "empty file",
			filePath:      emptyFile,
			expectError:   true,
			errorContains: "empty",
		},
		{
			name:          "non-existent file",
			filePath:      nonExistentFile,
			expectError:   true,
			errorContains: "does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFile(tt.filePath)

			if (err != nil) != tt.expectError {
				t.Errorf("validateFile() error = %v, expectError %v", err, tt.expectError)
			}

			if tt.expectError && err != nil && tt.errorContains != "" {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("validateFile() error = %v, should contain %q", err, tt.errorContains)
				}
			}
		})
	}
}

func TestFileValidator_AbsoluteAndRelativePaths(t *testing.T) {
	tmpDir := t.TempDir()

	readmePath := filepath.Join(tmpDir, "README.md")
	os.WriteFile(readmePath, []byte("# Test"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "outputs.tf"), []byte("output {}"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "variables.tf"), []byte("variable {}"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "terraform.tf"), []byte("terraform {}"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "relative.tf"), []byte("resource {}"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "absolute.tf"), []byte("data {}"), 0o644)

	tests := []struct {
		name            string
		additionalFiles []string
		expectedErrors  int
	}{
		{
			name:            "relative path gets converted",
			additionalFiles: []string{"relative.tf"},
			expectedErrors:  0,
		},
		{
			name:            "absolute path stays absolute",
			additionalFiles: []string{filepath.Join(tmpDir, "absolute.tf")},
			expectedErrors:  0,
		},
		{
			name:            "mixed relative and absolute",
			additionalFiles: []string{"relative.tf", filepath.Join(tmpDir, "absolute.tf")},
			expectedErrors:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fv := NewFileValidator(readmePath, tmpDir, tt.additionalFiles)
			errs := fv.Validate()

			if len(errs) != tt.expectedErrors {
				t.Errorf("Validate() returned %d errors; want %d", len(errs), tt.expectedErrors)
				for _, err := range errs {
					t.Logf("  error: %v", err)
				}
			}
		})
	}
}

func TestFileValidator_RequiredFilesContent(t *testing.T) {
	tmpDir := t.TempDir()

	readmePath := filepath.Join(tmpDir, "README.md")
	fv := NewFileValidator(readmePath, tmpDir, []string{})

	expectedFiles := []string{
		readmePath,
		filepath.Join(tmpDir, "outputs.tf"),
		filepath.Join(tmpDir, "variables.tf"),
		filepath.Join(tmpDir, "terraform.tf"),
	}

	for _, expected := range expectedFiles {
		found := false
		for _, actual := range fv.requiredFiles {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("NewFileValidator() missing required file %q", expected)
		}
	}
}

func TestFileValidator_ErrorMessages(t *testing.T) {
	tmpDir := t.TempDir()

	readmePath := filepath.Join(tmpDir, "README.md")
	os.WriteFile(readmePath, []byte("# Test"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "outputs.tf"), []byte("output {}"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "variables.tf"), []byte("variable {}"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "terraform.tf"), []byte{}, 0o644) // Empty file

	fv := NewFileValidator(readmePath, tmpDir, []string{"missing.tf"})
	errs := fv.Validate()

	if len(errs) != 2 {
		t.Errorf("Validate() returned %d errors; want 2", len(errs))
	}

	hasRequiredError := false
	hasAdditionalError := false

	for _, err := range errs {
		errMsg := err.Error()
		if strings.Contains(errMsg, "required") {
			hasRequiredError = true
		}
		if strings.Contains(errMsg, "additional") {
			hasAdditionalError = true
		}
	}

	if !hasRequiredError {
		t.Error("Expected at least one error message to contain 'required'")
	}

	if !hasAdditionalError {
		t.Error("Expected at least one error message to contain 'additional'")
	}
}
