package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShlex(t *testing.T) {
	s := `search "rick and morty" 1`
	res := shlex(s)

	require.Equal(t, "search", res[0])
	require.Equal(t, "rick and morty", res[1])
	require.Equal(t, "1", res[2])
}
