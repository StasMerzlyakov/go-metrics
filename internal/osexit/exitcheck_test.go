package osexit_test

import (
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/osexit"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestMyAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), osexit.Analyzer, "./...")

}
