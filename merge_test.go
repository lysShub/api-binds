package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestXxxx(t *testing.T) {

	// `D:\OneDrive\code\go\anton-planet-acceler\acceler\nodes\server\_binds\binds`

	s, err := merge(nil, "main")
	require.NoError(t, err)

	fh, err := os.Create("a.go")
	require.NoError(t, err)
	defer fh.Close()

	fh.WriteString(s)
}
