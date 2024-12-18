package license

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"iter"
	"sync"
)

// ID of a license. Must look like example de.worldiety.license.user.chat or de.worldiety.license.app.importer.
type ID string

func (id ID) Valid() bool {
	return permission.ID(id).Valid()
}

// AppLicense either just exists or not. It cannot be assigned to a user.
type AppLicense struct {
	ID          ID
	Name        string
	Description string `label:"Beschreibung"`
	Url         string `table-visible:"false" supportingText:"Dies wird u.a. als weiterführender Link in der Abrechnungsansicht dargestellt."`
	Incentive   string `table-visible:"false" label:"Incentive" supportingText:"Sofern gesetzt wird dies als 'jetzt anfragen' Anreiz-Link in der Abrechnung dargestellt."`
	Enabled     bool   `label:"Lizenz aktiv" supportingText:"nur aktive Lizenzen werden als verfügbar im System dargestellt."`
}

func (l AppLicense) WithIdentity(id ID) AppLicense {
	l.ID = id
	return l
}

func (l AppLicense) Identity() ID {
	return l.ID
}

// A UserLicense can be assigned to a single user and has an upper limit of how many users can be assigned in total.
// If MaxUsers is 0, it is unlimited.
// See also [AppLicense].
type UserLicense struct {
	ID          ID
	Name        string
	Description string `label:"Beschreibung"`
	Url         string `supportingText:"Dies wird u.a. als weiterführender Link in der Abrechnungsansicht dargestellt."`
	Incentive   string `label:"Incentive" supportingText:"Sofern gesetzt wird dies als 'jetzt anfragen' Anreiz-Link in der Abrechnung dargestellt."`
	MaxUsers    int    `label:"Maximale Anzahl an Nutzern bzw. zugeordneten Konten"`
}

func (l UserLicense) WithIdentity(id ID) UserLicense {
	l.ID = id
	return l
}

func (l UserLicense) Identity() ID {
	return l.ID
}

type FindAllAppLicenses func(subject permission.Auditable) iter.Seq2[AppLicense, error]
type FindAppLicenseByID func(subject permission.Auditable, id ID) (std.Option[AppLicense], error)
type CreateAppLicense func(subject permission.Auditable, license AppLicense) (ID, error)
type UpdateAppLicense func(subject permission.Auditable, license AppLicense) error
type DeleteAppLicense func(subject permission.Auditable, id ID) error

// UpsertAppLicense has a surprise comfort feature, because it does not overwrite the [AppLicense.Enabled] field.
type UpsertAppLicense func(subject permission.Auditable, license AppLicense) (ID, error)

type FindAllUserLicenses func(subject permission.Auditable) iter.Seq2[UserLicense, error]
type FindUserLicenseByID func(subject permission.Auditable, id ID) (std.Option[UserLicense], error)
type CreateUserLicense func(subject permission.Auditable, license UserLicense) (ID, error)
type UpdateUserLicense func(subject permission.Auditable, license UserLicense) error
type DeleteUserLicense func(subject permission.Auditable, id ID) error
type UpsertUserLicense func(subject permission.Auditable, license UserLicense) (ID, error)

type AppLicenseRepository data.Repository[AppLicense, ID]
type UserLicenseRepository data.Repository[UserLicense, ID]
type UseCases struct {
	PerUser struct {
		FindAll  FindAllUserLicenses
		FindByID FindUserLicenseByID
		Create   CreateUserLicense
		Update   UpdateUserLicense
		Delete   DeleteUserLicense
		Upsert   UpsertUserLicense
	}

	PerApp struct {
		FindAll  FindAllAppLicenses
		FindByID FindAppLicenseByID
		Create   CreateAppLicense
		Update   UpdateAppLicense
		Delete   DeleteAppLicense
		Upsert   UpsertAppLicense
	}
}

func NewUseCases(perAppRepo AppLicenseRepository, perUserRepo UserLicenseRepository) UseCases {
	// we are still in bootstrapping dependency cycle here, thus we cannot use the rcrud package

	mutex := new(sync.Mutex)
	findAllUserLicensesFn := NewFindAllUserLicenses(perUserRepo)
	findUserLicenseByIDFn := NewFindUserLicenseByID(perUserRepo)
	createUserLicenseFn := NewCreateUserLicense(mutex, perUserRepo)
	deleteUserLicenseFn := NewDeleteUserLicense(mutex, perUserRepo)
	updateUserLicenseFn := NewUpdateUserLicense(mutex, perUserRepo)
	upsertUserLicenceFn := NewUpsertUserLicense(mutex, perUserRepo)

	findAllAppLicensesFn := NewFindAllAppLicenses(perAppRepo)
	findAppLicenseByIDFn := NewFindAppLicenseByID(perAppRepo)
	createAppLicenseFn := NewCreateAppLicense(mutex, perAppRepo)
	deleteAppLicenseFn := NewDeleteAppLicense(mutex, perAppRepo)
	updateAppLicenseFn := NewUpdateAppLicense(mutex, perAppRepo)
	upsertAppLicenceFn := NewUpsertAppLicense(mutex, perAppRepo)

	var uc UseCases
	uc.PerUser.FindAll = findAllUserLicensesFn
	uc.PerUser.FindByID = findUserLicenseByIDFn
	uc.PerUser.Create = createUserLicenseFn
	uc.PerUser.Delete = deleteUserLicenseFn
	uc.PerUser.Update = updateUserLicenseFn
	uc.PerUser.Upsert = upsertUserLicenceFn

	uc.PerApp.FindAll = findAllAppLicensesFn
	uc.PerApp.FindByID = findAppLicenseByIDFn
	uc.PerApp.Create = createAppLicenseFn
	uc.PerApp.Delete = deleteAppLicenseFn
	uc.PerApp.Update = updateAppLicenseFn
	uc.PerApp.Upsert = upsertAppLicenceFn

	return uc
}
