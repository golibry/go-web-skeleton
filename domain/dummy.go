package domain

import "github.com/golibry/go-common-domain/domain"

type Dummy struct {
	id   int
	name string
}

func NewDummy(name string) (*Dummy, error) {
	dummy := &Dummy{}
	err := dummy.setName(name)
	return dummy, err
}

func ReconstituteDummy(id int, name string) *Dummy {
	return &Dummy{
		id:   id,
		name: name,
	}
}

func (d *Dummy) AddIdentity(id int) {
	d.id = id
}

func (d *Dummy) GetName() string {
	return d.name
}

func (d *Dummy) setName(name string) error {
	if len(name) == 0 {
		return domain.NewError("Name must not be empty")
	}

	d.name = name
	return nil
}

func (d *Dummy) GetId() int {
	return d.id
}
