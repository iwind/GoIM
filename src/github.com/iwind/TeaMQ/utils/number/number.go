package numberutil

func RangeInt(from int, to int, step uint) []int {
	if step == 0 {
		step = 1
	}
	var numbers []int
	var i int
	var intStep = int(step)
	if from < to {
		for i = from; i <= to; i += intStep {
			numbers = append(numbers, i)
		}
	} else {
		for i = from; i >= to; i -= intStep {
			numbers = append(numbers, i)
		}
	}
	return numbers
}

func Times(n uint, iterator func(i uint)) {
	var i uint
	for i = 0; i < n; i ++ {
		iterator(i)
	}
}
