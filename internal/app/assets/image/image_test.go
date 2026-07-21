package image

import "testing"

func TestAssetTypeIDMatchesImageAssetType(t *testing.T) {
	if assetTypeID != 1 {
		t.Fatalf("image assetTypeID = %d, want 1", assetTypeID)
	}
}
