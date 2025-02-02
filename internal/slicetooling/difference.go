package slicetooling

// Difference returns the difference between two slices of different types.
// Elements are compared using their respective ID functions.
// The first value is the list of elements in list1 which are absent in list2.
// The second value is the list of elements list2 which are absent in list1.
func Difference[T1 any, T2 any, C comparable](
	list1 []T1,
	list2 []T2,
	t1ID func(T1) C,
	t2ID func(T2) C,
) ([]T1, []T2) {
	left := []T1{}
	right := []T2{}

	seenLeft := map[C]struct{}{}
	seenRight := map[C]struct{}{}

	for _, elem := range list1 {
		seenLeft[t1ID(elem)] = struct{}{}
	}

	for _, elem := range list2 {
		seenRight[t2ID(elem)] = struct{}{}
	}

	for _, elem := range list1 {
		if _, ok := seenRight[t1ID(elem)]; !ok {
			left = append(left, elem)
		}
	}

	for _, elem := range list2 {
		if _, ok := seenLeft[t2ID(elem)]; !ok {
			right = append(right, elem)
		}
	}

	return left, right
}
