package response

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBatchMirrorKey(t *testing.T) {
	bk := batchMirrorKey{
		path:      "foo",
		companyID: "moov",
	}
	path := bk.getFilepath(nil)
	require.NotEmpty(t, path)
}
