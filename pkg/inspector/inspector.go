package inspector

type Inspector interface {
	// GetAll returns a map of all environment vars for a container.
	GetAllValues(containerId string) (map[string]string, error)

	// Returns the raw string value of the variable
	GetValue(containerId, varName string) (string, error)
}

func New() (Inspector, error) {
	return newDockerInspector()
}
