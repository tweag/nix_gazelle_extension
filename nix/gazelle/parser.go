package gazelle

type TraceOut struct {
	Cmd struct {
		Parent int
		ID     int
		Dir    string
		Path   string
		Args   []string
	}
	Inputs  []string
	Outputs []string
	FDs     struct {
		Num0 string
		Num1 string
		Num2 string
	}
}
