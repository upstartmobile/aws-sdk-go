// THIS FILE IS AUTOMATICALLY GENERATED. DO NOT EDIT.

package wafiface_test

import (
	"testing"

	"github.com/upstartmobile/aws-sdk-go/service/waf"
	"github.com/upstartmobile/aws-sdk-go/service/waf/wafiface"
	"github.com/stretchr/testify/assert"
)

func TestInterface(t *testing.T) {
	assert.Implements(t, (*wafiface.WAFAPI)(nil), waf.New(nil))
}
