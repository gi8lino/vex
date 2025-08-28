package bench_test

import (
	"os/exec"
	"testing"
)

// GNU envsubst (no operator semantics) → only the simple pattern.

func BenchmarkEnvsubst_OneSmallFile(b *testing.B) {
	b.ReportAllocs()
	if _, err := exec.LookPath(envsubstBin()); err != nil {
		b.Skip("envsubst not found in PATH (override with ENV_BIN)")
	}
	dir := b.TempDir()
	files, total := makeFiles(b, dir, 1, BenchSmallFileSize, BenchEnvsubstPattern)

	b.SetBytes(total)
	b.ResetTimer()

	for b.Loop() {
		if err := runStreaming(envsubstBin(), nil, files, BenchEnvVars); err != nil {
			b.Fatalf("envsubst: %v", err)
		}
	}
}

func BenchmarkEnvsubst_ManySmallFiles(b *testing.B) {
	b.ReportAllocs()
	if _, err := exec.LookPath(envsubstBin()); err != nil {
		b.Skip("envsubst not found in PATH (override with ENV_BIN)")
	}
	dir := b.TempDir()
	files, total := makeFiles(b, dir, BenchNumSmallFiles, BenchSmallFileSize, BenchEnvsubstPattern)

	b.SetBytes(total)
	b.ResetTimer()

	for b.Loop() {
		if err := runStreaming(envsubstBin(), nil, files, BenchEnvVars); err != nil {
			b.Fatalf("envsubst: %v", err)
		}
	}
}

func BenchmarkEnvsubst_OneBigFile(b *testing.B) {
	b.ReportAllocs()
	if _, err := exec.LookPath(envsubstBin()); err != nil {
		b.Skip("envsubst not found in PATH (override with ENV_BIN)")
	}
	dir := b.TempDir()
	files, total := makeFiles(b, dir, 1, BenchBigFileSize, BenchEnvsubstPattern)

	b.SetBytes(total)
	b.ResetTimer()

	for b.Loop() {
		if err := runStreaming(envsubstBin(), nil, files, BenchEnvVars); err != nil {
			b.Fatalf("envsubst: %v", err)
		}
	}
}

func BenchmarkEnvsubst_ManyBigFiles(b *testing.B) {
	b.ReportAllocs()
	if _, err := exec.LookPath(envsubstBin()); err != nil {
		b.Skip("envsubst not found in PATH (override with ENV_BIN)")
	}
	dir := b.TempDir()
	files, total := makeFiles(b, dir, BenchNumBigFiles, BenchBigFileSize, BenchEnvsubstPattern)

	b.SetBytes(total)
	b.ResetTimer()

	for b.Loop() {
		if err := runStreaming(envsubstBin(), nil, files, BenchEnvVars); err != nil {
			b.Fatalf("envsubst: %v", err)
		}
	}
}
