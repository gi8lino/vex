package flag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFlags(t *testing.T) {
	t.Parallel()

	t.Run("no args", func(t *testing.T) {
		t.Parallel()

		flags, err := ParseFlags([]string{}, "1.2.3", "abc123")
		require.NoError(t, err)

		assert.False(t, flags.InPlace)
		assert.Equal(t, "", flags.BackupExt)

		assert.False(t, flags.NoOps)
		assert.False(t, flags.NoEscape)

		assert.False(t, flags.ErrorEmpty)
		assert.False(t, flags.ErrorUnset)

		assert.False(t, flags.KeepUnset)
		assert.False(t, flags.KeepEmpty)
		assert.False(t, flags.KeepVars) // not set by parser; remains default false

		assert.Empty(t, flags.Prefix)
		assert.Empty(t, flags.Suffix)
		assert.Empty(t, flags.Variables)

		assert.False(t, flags.Colored)

		assert.Empty(t, flags.Positional)
	})

	t.Run("in-place and backup", func(t *testing.T) {
		t.Parallel()
		flags, err := ParseFlags([]string{
			"-i",
			"--backup", ".bak",
		}, "1.0.0", "deadbeef")
		require.NoError(t, err)

		assert.True(t, flags.InPlace)
		assert.Equal(t, ".bak", flags.BackupExt)
	})

	t.Run("Backup without --in-place", func(t *testing.T) {
		t.Parallel()
		flags, err := ParseFlags([]string{
			"--backup", ".bak",
		}, "1.0.0", "deadbeef")
		require.Error(t, err)
		assert.EqualError(t, err, "--backup requires --in-place")
		assert.Empty(t, flags)
	})

	t.Run("BackupExt with leading dot", func(t *testing.T) {
		t.Parallel()
		flags, err := ParseFlags([]string{
			"--in-place",
			"--backup", ".bak",
		}, "1.0.0", "deadbeef")
		require.NoError(t, err)
		assert.Equal(t, ".bak", flags.BackupExt)
	})

	t.Run("BackupExt without leading dot", func(t *testing.T) {
		t.Parallel()
		flags, err := ParseFlags([]string{
			"--in-place",
			"--backup", "bak",
		}, "1.0.0", "deadbeef")
		require.NoError(t, err)
		assert.Equal(t, ".bak", flags.BackupExt)
	})

	t.Run("no-ops", func(t *testing.T) {
		t.Parallel()

		flags, err := ParseFlags([]string{"--no-ops", "--literal-dollar"}, "1.0.0", "deadbeef")
		require.NoError(t, err)

		assert.True(t, flags.NoOps)
		assert.True(t, flags.NoEscape)
	})

	t.Run("strict implies both error flags", func(t *testing.T) {
		t.Parallel()

		flags, err := ParseFlags([]string{"--strict"}, "1.0.0", "deadbeef")
		require.NoError(t, err)

		assert.True(t, flags.ErrorUnset)
		assert.True(t, flags.ErrorEmpty)
	})

	t.Run("keep-vars implies keep-unset and keep-empty", func(t *testing.T) {
		t.Parallel()

		flags, err := ParseFlags([]string{"--keep-vars"}, "1.0.0", "deadbeef")
		require.NoError(t, err)

		// --keep-vars implies both keep-* flags
		assert.True(t, flags.KeepUnset)
		assert.True(t, flags.KeepEmpty)

		// Note: Options.NoReplace is not set by ParseFlags; it remains false by design here.
		assert.False(t, flags.KeepVars)
	})

	t.Run("prefix filters", func(t *testing.T) {
		t.Parallel()

		args := []string{
			"--prefix", "APP_",
			"-p", "SYS_",
			"--suffix", "_TOKEN",
			"-s", "_ID",
			"--variable", "FOO",
			"-v", "BAR",
		}
		flags, err := ParseFlags(args, "1.0.0", "deadbeef")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"APP_", "SYS_"}, flags.Prefix)
		assert.ElementsMatch(t, []string{"_TOKEN", "_ID"}, flags.Suffix)
		assert.ElementsMatch(t, []string{"FOO", "BAR"}, flags.Variables)
	})

	t.Run("colored", func(t *testing.T) {
		t.Parallel()

		flags, err := ParseFlags([]string{"--colored"}, "1.0.0", "deadbeef")
		require.NoError(t, err)

		assert.True(t, flags.Colored)
	})

	t.Run("positional args", func(t *testing.T) {
		t.Parallel()

		flags, err := ParseFlags([]string{
			"--no-ops",
			"file1.txt",
			"file two.md",
		}, "1.0.0", "deadbeef")
		require.NoError(t, err)

		assert.True(t, flags.NoOps)
		require.Len(t, flags.Positional, 2)
		assert.Equal(t, "file1.txt", flags.Positional[0])
		assert.Equal(t, "file two.md", flags.Positional[1])
	})

	t.Run("invalid args", func(t *testing.T) {
		t.Parallel()
		flags, err := ParseFlags([]string{
			"--invalid-flag",
		}, "1.0.0", "deadbeef")
		require.Error(t, err)
		assert.EqualError(t, err, "unknown flag: --invalid-flag")
		assert.Empty(t, flags)
	})
}
