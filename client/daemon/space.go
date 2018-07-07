package daemon

import "fmt"

// Space user space
type Space struct {
	SpaceNo    uint32
	Password   string
	EncryptKey []byte
	Root       string
	Name       string
}

// NewSpace create user space
func NewSpace(no uint32, password string, root string) *Space {
	return &Space{
		SpaceNo:    no,
		Password:   password,
		EncryptKey: []byte(password),
		Root:       root,
	}
}

// SpaceManager manage multi user space
type SpaceManager struct {
	AS      []*Space
	Current uint32
	Count   uint32
}

func NewSpaceManager() *SpaceManager {
	return &SpaceManager{
		AS:      make([]*Space, 0, 1),
		Current: 0,
		Count:   0,
	}
}

// AddSpace add a user space
func (m *SpaceManager) AddSpace(no uint32, password, root string) {
	m.AS = append(m.AS, NewSpace(no, password, root))
	m.Count++
}

// Switch switch space
func (m *SpaceManager) Switch(no uint32) error {
	if no >= m.Count {
		return fmt.Errorf("space %d not exists", no)
	}
	m.Current = no
	return nil
}

// GetSpacePasswd return passord of space no
func (m *SpaceManager) GetSpacePasswd(no uint32) ([]byte, error) {
	if no >= m.Count {
		return nil, fmt.Errorf("space %d not exists", no)
	}

	return m.AS[no].EncryptKey, nil
}

// SetSpacePasswd  set passord of some space
func (m *SpaceManager) SetSpacePasswd(no uint32, password string) error {
	if no >= m.Count {
		return fmt.Errorf("space %d not exists", no)
	}

	m.AS[no].Password = password
	m.AS[no].EncryptKey = []byte(password)
	return nil
}
