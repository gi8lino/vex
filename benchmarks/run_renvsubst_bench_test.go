package bench_test

import (
	"os/exec"
	"testing"
)

func BenchmarkRenvsubst_OneSmallFile(b *testing.B) {
	b.ReportAllocs()
	if _, err := exec.LookPath(renvsubstBin()); err != nil {
		b.Skip("renvsubst not found (set RENV_BIN or put ./bin/renvsubst in place)")
	}
	dir := b.TempDir()
	files, total := makeFiles(b, dir, 1, BenchSmallFileSize, BenchExtendedPattern)

	b.SetBytes(total)
	b.ResetTimer()

	for b.Loop() {
		if err := runStreaming(renvsubstBin(), nil, files, BenchEnvVars); err != nil {
			b.Fatalf("renvsubst: %v", err)
		}
	}
}

func BenchmarkRenvsubst_ManySmallFiles(b *testing.B) {
	b.ReportAllocs()
	if _, err := exec.LookPath(renvsubstBin()); err != nil {
		b.Skip("renvsubst not found (set RENV_BIN or put ./bin/renvsubst in place)")
	}
	dir := b.TempDir()
	files, total := makeFiles(b, dir, BenchNumSmallFiles, BenchSmallFileSize, BenchExtendedPattern)

	b.SetBytes(total)
	b.ResetTimer()

	for b.Loop() {
		if err := runStreaming(renvsubstBin(), nil, files, BenchEnvVars); err != nil {
			b.Fatalf("renvsubst: %v", err)
		}
	}
}

func BenchmarkRenvsubst_OneBigFile(b *testing.B) {
	b.ReportAllocs()
	if _, err := exec.LookPath(renvsubstBin()); err != nil {
		b.Skip("renvsubst not found (set RENV_BIN or put ./bin/renvsubst in place)")
	}
	dir := b.TempDir()
	files, total := makeFiles(b, dir, 1, BenchBigFileSize, BenchExtendedPattern)

	b.SetBytes(total)
	b.ResetTimer()

	for b.Loop() {
		if err := runStreaming(renvsubstBin(), nil, files, BenchEnvVars); err != nil {
			b.Fatalf("renvsubst: %v", err)
		}
	}
}

func BenchmarkRenvsubst_ManyBigFiles(b *testing.B) {
	b.ReportAllocs()
	if _, err := exec.LookPath(renvsubstBin()); err != nil {
		b.Skip("renvsubst not found (set RENV_BIN or put ./bin/renvsubst in place)")
	}
	dir := b.TempDir()
	files, total := makeFiles(b, dir, BenchNumBigFiles, BenchBigFileSize, BenchExtendedPattern)

	b.SetBytes(total)
	b.ResetTimer()

	for b.Loop() {
		if err := runStreaming(renvsubstBin(), nil, files, BenchEnvVars); err != nil {
			b.Fatalf("renvsubst: %v", err)
		}
	}
}

func BenchmarkRenvsubst_ExtendedCases(b *testing.B) {
	if _, err := exec.LookPath(renvsubstBin()); err != nil {
		b.Skip("renvsubst not found (set RENV_BIN or put ./bin/renvsubst in place)")
	}
	// If you want to filter unsupported ops (case transforms), you can reuse the earlier filter.
	for _, tc := range extendedCases {
		// Example skip: if the tool lacks case ops, skip those tests.
		skip := false
		for _, op := range tc.ops {
			switch op {
			case OpCaseFirstUp, OpCaseAllUp, OpCaseFirstLo, OpCaseAllLo:
				skip = true
			}
		}
		if skip {
			continue
		}

		b.Run(tc.name, func(b *testing.B) {
			dir := b.TempDir()
			files, _ := makeFiles(b, dir, BenchNumSmallFiles, 8*1024, tc.pattern)
			b.ReportAllocs()
			for b.Loop() {
				if err := runStreaming(renvsubstBin(), nil, files, tc.env); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
