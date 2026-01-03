package markparsr

type ErrorCollector struct {
	errs []error
}

func (c *ErrorCollector) Add(err error) {
	if err != nil {
		c.errs = append(c.errs, err)
	}
}

func (c *ErrorCollector) AddMany(errs []error) {
	for _, err := range errs {
		c.Add(err)
	}
}

func (c *ErrorCollector) Errors() []error {
	return c.errs
}

func (c *ErrorCollector) HasErrors() bool {
	return len(c.errs) > 0
}
