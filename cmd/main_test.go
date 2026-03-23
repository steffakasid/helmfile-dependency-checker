package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/steffenrumpf/hdc/internal/models"
)

func TestClassifyExitCode_Clean(t *testing.T) {
	result := &models.Result{
		Findings: []models.Finding{
			{Status: models.StatusOK},
			{Status: models.StatusSkipped},
		},
	}
	assert.Equal(t, 0, classifyExitCode(result))
}

func TestClassifyExitCode_WarningsOnly(t *testing.T) {
	result := &models.Result{
		Findings: []models.Finding{
			{Status: models.StatusOK},
			{Status: models.StatusOutdated},
			{Status: models.StatusSkipped},
		},
	}
	assert.Equal(t, 1, classifyExitCode(result))
}

func TestClassifyExitCode_ErrorsPresent(t *testing.T) {
	result := &models.Result{
		Findings: []models.Finding{
			{Status: models.StatusOutdated},
			{Status: models.StatusUnmaintained},
			{Status: models.StatusOK},
		},
	}
	assert.Equal(t, 2, classifyExitCode(result))
}

func TestClassifyExitCode_UnreachableIsError(t *testing.T) {
	result := &models.Result{
		Findings: []models.Finding{
			{Status: models.StatusUnreachable},
		},
	}
	assert.Equal(t, 2, classifyExitCode(result))
}

func TestClassifyExitCode_SkippedNeverAffectsCode(t *testing.T) {
	result := &models.Result{
		Findings: []models.Finding{
			{Status: models.StatusSkipped},
			{Status: models.StatusSkipped},
		},
	}
	assert.Equal(t, 0, classifyExitCode(result))
}

func TestClassifyExitCode_Empty(t *testing.T) {
	result := &models.Result{}
	assert.Equal(t, 0, classifyExitCode(result))
}

func TestExitError(t *testing.T) {
	err := &exitError{code: 2, message: "test error"}
	assert.Equal(t, "test error", err.Error())
	assert.Equal(t, 2, err.code)
}
