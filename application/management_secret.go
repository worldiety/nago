package application

import (
	"fmt"
	"go.wdy.de/nago/application/secret"
	uisecret "go.wdy.de/nago/application/secret/ui"
	"go.wdy.de/nago/pkg/blob/crypto"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
)

type SecretManagement struct {
	UseCases secret.UseCases
	Pages    uisecret.Pages
}

func (c *Configurator) SecretManagement() (SecretManagement, error) {
	if c.secretManagement == nil {

		users, err := c.UserManagement()
		if err != nil {
			return SecretManagement{}, err
		}

		groups, err := c.GroupManagement()
		if err != nil {
			return SecretManagement{}, err
		}

		key, err := c.MasterKey()
		if err != nil {
			return SecretManagement{}, fmt.Errorf("could not load master key: %w", err)
		}

		secretStore, err := c.EntityStore("nago.iam.secret")
		if err != nil {
			return SecretManagement{}, fmt.Errorf("cannot get entity store: %w", err)
		}

		encryptedSecretStore := crypto.NewBlobStore(secretStore, key)
		secretRepo := json.NewSloppyJSONRepository[secret.Secret, secret.ID](encryptedSecretStore)
		uc := secret.NewUseCases(secretRepo)

		c.secretManagement = &SecretManagement{
			UseCases: uc,
			Pages: uisecret.Pages{
				Vault:        "admin/secret/vault",
				CreateSecret: "admin/secret/create",
				EditSecret:   "admin/secret/edit",
			},
		}

		c.RootViewWithDecoration(c.secretManagement.Pages.Vault, func(wnd core.Window) core.View {
			return uisecret.VaultPage(wnd, c.secretManagement.Pages, uc.FindMySecrets, groups.UseCases.FindByID)
		})

		c.RootViewWithDecoration(c.secretManagement.Pages.CreateSecret, func(wnd core.Window) core.View {
			return uisecret.CreateSecretPage(wnd, c.secretManagement.Pages, uc.CreateSecret)
		})

		c.RootViewWithDecoration(c.secretManagement.Pages.EditSecret, func(wnd core.Window) core.View {
			return uisecret.EditSecretPage(
				wnd,
				c.secretManagement.Pages,
				uc.DeleteMySecretByID,
				uc.FindMySecretByID,
				uc.UpdateMyCredentials,
				uc.UpdateMySecretGroups,
				uc.UpdateMySecretOwners,
				groups.UseCases.FindMyGroups,
				users.UseCases.FindAll,
			)
		})
	}

	return *c.secretManagement, nil
}
