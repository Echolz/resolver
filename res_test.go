package resolver

import (
	"testing"
)

//, nil

func BenchmarkSomething(b *testing.B) {
	fPersonName := "somename1"
	sPersonName := "somename2"
	expression := "user.Parents[0].Parents[0].Parent.Name"

	user := Person{
		Name: &fPersonName,
		Parent: &Person{
			Name: &sPersonName,
		},
		Parents: []*Person{
			{
				Name: &sPersonName,
				Parent: &Person{
					Name: &fPersonName,
				},
				Parents: []*Person{
					{
						Name: &sPersonName,
						Parent: &Person{
							Name: &fPersonName,
						},
					},
				},
			},
		},
	}

	resolverArgs := map[string]interface{}{
		"user": user,
	}

	r := NewResolver(resolverArgs)

	for i := 0; i < b.N; i++ {
		v, err := r.Resolve(expression)
		if err != nil {
			b.Error("Expected nil error")
		}

		if v == nil {
			b.Error("Expected nil error")
		}
	}
}
