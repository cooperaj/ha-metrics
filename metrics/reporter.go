package metrics

type reporter interface {
	Report(string, interface{})
}
