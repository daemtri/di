package di

var (
	validateFunc = func(x any) error {
		return nil
	}
)

func SetValidator(fn func(x any) error) {
	validateFunc = fn
}
