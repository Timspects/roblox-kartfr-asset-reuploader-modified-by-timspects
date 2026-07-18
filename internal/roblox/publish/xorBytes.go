package publish

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Timspects/roblox-kartfr-asset-reuploader-modified-by-timspects/internal/roblox"
)

func xorBytes(data []byte, key byte) []byte {
	out := make([]byte, len(data))
	for i := range data {
		out[i] = data[i] ^ key
	}
	return out
}
