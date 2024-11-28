package license

import (
	"go.wdy.de/nago/pkg/data/mem"
	"regexp"
)

var Global = &mem.Repository[License, ID]{}

var regexPermissionID = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z0-9_]+)*[a-z0-9_]*$`)

// ID of a license. Must look like example de.worldiety.license.user.chat or de.worldiety.license.app.importer.
type ID string

func (id ID) Valid() bool {
	return regexPermissionID.MatchString(string(id))
}

// License if the common type of all available licenses. See also [UserLicense] and [AppLicense].
type License interface {
	Identity() ID
	LicenseName() string
	Enabled() bool
}

// A UserLicense can be assigned to a single user and has an upper limit of how many users can be assigned.
// If MaxUsers is 0, it is unlimited.
// See also [AppLicense].
type UserLicense struct {
	ID          ID
	Name        string
	Description string
	Url         string
	MaxUsers    int
	IsEnabled   bool
}

func (l UserLicense) Enabled() bool {
	return l.IsEnabled
}

func (l UserLicense) Identity() ID {
	return l.ID
}

func (l UserLicense) LicenseName() string {
	return l.Name
}

func (l UserLicense) Unlimited() bool {
	return l.MaxUsers == 0
}

// AppLicense either just exists or not. It cannot be assigned to a user.
type AppLicense struct {
	ID          ID
	Name        string
	Description string
	Url         string
	IsEnabled   bool
}

func (l AppLicense) Identity() ID {
	return l.ID
}

func (l AppLicense) Enabled() bool {
	return l.IsEnabled
}

func (l AppLicense) LicenseName() string {
	return l.Name
}

type Status struct {
	Licenses     []License
	AppliedUsers map[ID]User
}

type User interface {
	ID() string // cannot use auth.UID due to package cycle
	Firstname() string
	Lastname() string
	HasLicense(id ID) bool
}

/*
var permCalculateStatus = annotation.Permission[CalculateStatus]("de.worldiety.nago.license.status")

type CalculateStatus func(subject auth.Subject) (Status, error)

func NewCalculateStatus(users iter.Seq2[User, error], licenses iter.Seq2[License, error]) CalculateStatus {
	return func(subject auth.Subject) (Status, error) {
		if err := subject.Audit(permCalculateStatus.Identity()); err != nil {
			return Status{}, err
		}

	}
}*/
