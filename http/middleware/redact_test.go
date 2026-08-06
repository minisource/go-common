package middleware

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedactJSON_NestedAndCaseInsensitive(t *testing.T) {
	in := []byte(`{
		"user": {
			"Password": "s3cret",
			"email": "a@b.c",
			"meta": {"OTP": "123456", "note": "keep me"}
		},
		"token": "abc",
		"safe": "visible"
	}`)
	out := RedactJSON(in, DefaultRedactFields())
	require.NotNil(t, out)

	var v map[string]interface{}
	require.NoError(t, json.Unmarshal(out, &v))
	assert.Equal(t, RedactedValue, v["token"])
	assert.Equal(t, "visible", v["safe"])

	user := v["user"].(map[string]interface{})
	assert.Equal(t, RedactedValue, user["Password"])
	assert.Equal(t, "a@b.c", user["email"])
	meta := user["meta"].(map[string]interface{})
	assert.Equal(t, RedactedValue, meta["OTP"])
	assert.Equal(t, "keep me", meta["note"])
}

func TestRedactJSON_Arrays(t *testing.T) {
	in := []byte(`{"items":[{"password":"x"},{"token":"y","name":"ok"}],"secret":"z"}`)
	out := RedactJSON(in, DefaultRedactFields())
	require.NotNil(t, out)

	var v map[string]interface{}
	require.NoError(t, json.Unmarshal(out, &v))
	assert.Equal(t, RedactedValue, v["secret"])
	items := v["items"].([]interface{})
	assert.Equal(t, RedactedValue, items[0].(map[string]interface{})["password"])
	assert.Equal(t, RedactedValue, items[1].(map[string]interface{})["token"])
	assert.Equal(t, "ok", items[1].(map[string]interface{})["name"])
}

func TestRedactJSON_MalformedReturnsNil(t *testing.T) {
	assert.Nil(t, RedactJSON([]byte(`{"unterminated`), DefaultRedactFields()))
	assert.Nil(t, RedactJSON([]byte(`not json at all`), DefaultRedactFields()))
	assert.Nil(t, RedactJSON(nil, DefaultRedactFields()))
}

func TestRedactJSON_SensitiveValueNotRecursed(t *testing.T) {
	// A redacted value must be replaced wholesale — nested secrets cannot leak.
	in := []byte(`{"token":{"password":"leak","nested":{"otp":"123"}}}`)
	out := RedactJSON(in, DefaultRedactFields())
	require.NotNil(t, out)

	var v map[string]interface{}
	require.NoError(t, json.Unmarshal(out, &v))
	assert.Equal(t, RedactedValue, v["token"])
	assert.NotContains(t, string(out), "leak")
	assert.NotContains(t, string(out), "123")
}
