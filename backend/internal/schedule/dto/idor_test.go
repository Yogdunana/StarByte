package dto

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListRequestsRejectOwnerID(t *testing.T) {
	for _, typ := range []reflect.Type{
		reflect.TypeOf(ListCalendarRequest{}),
		reflect.TypeOf(ListEventRequest{}),
		reflect.TypeOf(RangeEventRequest{}),
	} {
		for i := 0; i < typ.NumField(); i++ {
			name := strings.ToLower(typ.Field(i).Name)
			tag := strings.ToLower(typ.Field(i).Tag.Get("form") + typ.Field(i).Tag.Get("json"))
			assert.NotContains(t, name, "owner", "%s must not expose owner filter", typ.Name())
			assert.NotContains(t, tag, "owner_id", "%s must not bind owner_id", typ.Name())
		}
	}
}
