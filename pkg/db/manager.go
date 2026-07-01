package db

type Manager struct {
	Connection
}

func New(connector Connection) *Manager {
	return &Manager{
		Connection: connector,
	}
}
