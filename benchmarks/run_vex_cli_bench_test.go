package bench_test

import (
	"os/exec"
	"testing"
)

// ---------------- Vex CLI (streams stdin->stdout) -----------------

func BenchmarkVexCLI_OneSmallFile(b *testing.B) {
	b.ReportAllocs()
	if _, err := exec.LookPath(vexBin()); err != nil {
		b.Skip("vex not found (set VEX_BIN)")
	}
	dir := b.TempDir()
	files, total := makeFiles(b, dir, 1, BenchSmallFileSize, BenchExtendedPattern)

	b.SetBytes(total)
	b.ResetTimer()

	for b.Loop() {
		if err := runStreaming(vexBin(), vexArgs(), files, BenchEnvVars); err != nil {
			b.Fatalf("vex cli: %v", err)
		}
	}
}

func BenchmarkVexCLI_ManySmallFiles(b *testing.B) {
	b.ReportAllocs()
	if _, err := exec.LookPath(vexBin()); err != nil {
		b.Skip("vex not found (set VEX_BIN)")
	}
	dir := b.TempDir()
	files, total := makeFiles(b, dir, BenchNumSmallFiles, BenchSmallFileSize, BenchExtendedPattern)

	b.SetBytes(total)
	b.ResetTimer()

	for b.Loop() {
		if err := runStreaming(vexBin(), vexArgs(), files, BenchEnvVars); err != nil {
			b.Fatalf("vex cli: %v", err)
		}
	}
}

func BenchmarkVexCLI_OneBigFile(b *testing.B) {
	b.ReportAllocs()
	if _, err := exec.LookPath(vexBin()); err != nil {
		b.Skip("vex not found (set VEX_BIN)")
	}
	dir := b.TempDir()
	files, total := makeFiles(b, dir, 1, BenchBigFileSize, BenchExtendedPattern)

	b.SetBytes(total)
	b.ResetTimer()

	for b.Loop() {
		if err := runStreaming(vexBin(), vexArgs(), files, BenchEnvVars); err != nil {
			b.Fatalf("vex cli: %v", err)
		}
	}
}

func BenchmarkVexCLI_ManyBigFiles(b *testing.B) {
	b.ReportAllocs()
	if _, err := exec.LookPath(vexBin()); err != nil {
		b.Skip("vex not found (set VEX_BIN)")
	}
	dir := b.TempDir()
	files, total := makeFiles(b, dir, BenchNumBigFiles, BenchBigFileSize, BenchExtendedPattern)

	b.SetBytes(total)
	b.ResetTimer()

	for b.Loop() {
		if err := runStreaming(vexBin(), vexArgs(), files, BenchEnvVars); err != nil {
			b.Fatalf("vex cli: %v", err)
		}
	}
}

// --- Vex CLI: Extended op matrix (per-case; non-amortized to match your prior runs) ---

func BenchmarkVexCLI_ExtendedCases(b *testing.B) {
	if _, err := exec.LookPath(vexBin()); err != nil {
		b.Skip("vex not found (set VEX_BIN)")
	}
	for _, tc := range extendedCases {
		b.Run(tc.name, func(b *testing.B) {
			dir := b.TempDir()
			files, _ := makeFiles(b, dir, BenchNumSmallFiles, 8*1024, tc.pattern)
			env := tc.env
			b.ReportAllocs()
			for b.Loop() {
				if err := runStreaming(vexBin(), vexArgs(), files, env); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
