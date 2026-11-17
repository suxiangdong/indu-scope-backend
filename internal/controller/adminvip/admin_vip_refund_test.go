package adminvip

import (
	"reflect"
	"testing"
)

func TestRefundMutationsAreNotBoundToOrdinaryAdminController(t *testing.T) {
	ordinary, privileged := reflect.TypeOf(NewV1()), reflect.TypeOf(NewRefundV1())
	for _, name := range []string{"CreateRefund", "RetryRefund"} {
		if _, exists := ordinary.MethodByName(name); exists {
			t.Fatalf("%s would be exposed by ordinary controller binding", name)
		}
		if _, exists := privileged.MethodByName(name); !exists {
			t.Fatalf("%s is missing from privileged controller", name)
		}
	}
	for _, name := range []string{"ListOrders", "ListRefunds", "Refund"} {
		if _, exists := ordinary.MethodByName(name); !exists {
			t.Fatalf("ordinary administrators lost read access to %s", name)
		}
	}
}
