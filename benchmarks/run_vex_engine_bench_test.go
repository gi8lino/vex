package bench_test

import (
	"io"
	"strings"
	"testing"

	"github.com/gi8lino/vex/internal/app"
)

// ---------------- Vex internal (direct app.Run) -----------------

// benchLookup provides variable lookup for ${A:-a} and ${B}.
func benchLookup(name string) (string, bool) {
	switch name {
	case "B":
		return "bee", true
	case "A":
		return "", false // unset -> default applies
	default:
		return "", false
	}
}

func BenchmarkVexRun_OneSmallFile(b *testing.B) {
	b.ReportAllocs()
	dir := b.TempDir()
	files, total := makeFiles(b, dir, 1, BenchSmallFileSize, BenchExtendedPattern)

	b.SetBytes(total)
	b.ResetTimer()

	for b.Loop() {
		out := io.Discard
		err := app.Run("v", "c", files, out, strings.NewReader(""), benchLookup, nil)
		if err != nil {
			b.Fatalf("Run err=%v", err)
		}
	}
}

func BenchmarkVexRun_ManySmallFiles(b *testing.B) {
	b.ReportAllocs()
	dir := b.TempDir()
	files, total := makeFiles(b, dir, BenchNumSmallFiles, BenchSmallFileSize, BenchExtendedPattern)

	b.SetBytes(total)
	b.ResetTimer()

	for b.Loop() {
		out := io.Discard
		err := app.Run("v", "c", files, out, strings.NewReader(""), benchLookup, nil)
		if err != nil {
			b.Fatalf("Run err=%v", err)
		}
	}
}

func BenchmarkVexRun_OneBigFile(b *testing.B) {
	b.ReportAllocs()
	dir := b.TempDir()
	files, total := makeFiles(b, dir, 1, BenchBigFileSize, BenchExtendedPattern)

	b.SetBytes(total)
	b.ResetTimer()

	for b.Loop() {
		out := io.Discard
		err := app.Run("v", "c", files, out, strings.NewReader(""), benchLookup, nil)
		if err != nil {
			b.Fatalf("Run err=%v", err)
		}
	}
}

func BenchmarkVexRun_ManyBigFiles(b *testing.B) {
	b.ReportAllocs()
	dir := b.TempDir()
	files, total := makeFiles(b, dir, BenchNumBigFiles, BenchBigFileSize, BenchExtendedPattern)

	b.SetBytes(total)
	b.ResetTimer()

	for b.Loop() {
		out := io.Discard
		err := app.Run("v", "c", files, out, strings.NewReader(""), benchLookup, nil)
		if err != nil {
			b.Fatalf("Run err=%v", err)
		}
	}
}
