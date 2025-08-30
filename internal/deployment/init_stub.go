//go:build !wireinject

package deployment

// Init provides a no-op implementation for tests.
func Init(configuration ApplicationConfiguration) Application {
	return Application{}
}
