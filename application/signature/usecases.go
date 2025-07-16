// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package signature

import (
	"crypto/sha3"
	"encoding/hex"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xtime"
	"hash"
	"iter"
	"strconv"
	"sync"
)

type ID string

// Sha3H256 contains the hex encoded sha3 256 hash of the referenced data bytes. For JSON or plain text this
// refers usually to the uncompressed data set. Images are hashed on an as-is basis of the blob.
type Sha3H256 string

// Signature represents (currently) an unqualified electronic signature. This may be extended by additional
// fields in the future to add TSA support or even fully qualified certificates. However, to protect against
// manipulations, all our signatures form a hash-chain, based on absolute sign order.
//
// A Signature is also immutable if used as a value type. Mutations on one copy cannot accidentally
// mutate other copies, and this property must be respected for future extensions.
type Signature struct {
	ID ID `json:"id,omitempty"`

	// Hash is the internal signature hash based on all fields below. It is calculated on
	// Hash = sha3( No | PreviousSignatureHash | unix-timestamp | Timezone | Firstname | Lastname | User | Email | ImageID | ImageHash | Documents)
	Hash Sha3H256 `json:"hash,omitempty"`

	// hash everything below

	// Number is a strict monotonic incrementing counter. It is never <= 0.
	Number int `json:"number,omitempty"`

	PreviousSignatureHash Sha3H256 `json:"previousSignatureHash"`

	// Timestamp in unix milliseconds when this signature has been created.
	Timestamp xtime.UnixMilliseconds `json:"timestamp,omitempty"`

	// Timezone in which this signature was created as IANA TZDB Identifier e.g. Europe/Berlin.
	Timezone string `json:"timezone,omitempty"`

	// Firstname of the person who created this signature. Must never be empty.
	Firstname string `json:"firstname"`

	// Lastname of the person who created this signature. Must never be empty.
	Lastname string `json:"lastname"`

	// User is an optional user identifier from the system which links to the creator of the authenticated user
	// when this signature was created.
	User user.ID `json:"user,omitempty"`

	// Email is optional and may be provided beside the name.
	Email user.Email `json:"email,omitempty"`

	// Image is an optional image identifier and may refer to an unqualified image representation of a
	// handwritten signature.
	Image image.ID `json:"image,omitempty"`

	// ImageSha3H256 is the hash of the unqualified image content. If image is empty, the hash is also the empty string.
	ImageHash Sha3H256 `json:"sha3H256,omitempty"`

	// Documents contains the meta data about the signed documents.
	Documents xslices.Slice[Document] `json:"documents,omitempty"`
}

func (s Signature) Identity() ID {
	return s.ID
}

func (s Signature) CalcHash() Sha3H256 {
	h := sha3.New256()
	w(h, strconv.Itoa(s.Number))
	w(h, s.PreviousSignatureHash)
	w(h, strconv.FormatInt(int64(s.Timestamp), 10))
	w(h, s.Timezone)
	w(h, s.Firstname)
	w(h, s.Lastname)
	w(h, s.User)
	w(h, s.Image)
	w(h, s.ImageHash)
	for document := range s.Documents.All() {
		w(h, document.Resource.Name)
		w(h, document.Resource.ID)
		w(h, strconv.FormatInt(document.Size, 10))
		w(h, document.Hash)
	}

	sum := h.Sum(nil)
	return Sha3H256(hex.EncodeToString(sum))
}

func w[Str ~string](h hash.Hash, str Str) {
	if _, err := h.Write([]byte(str)); err != nil {
		panic(err)
	}
}

type Document struct {
	// Resource references any internal resolvable resource.
	Resource user.Resource

	// Size in bytes of the data bytes.
	Size int64 `json:"size,omitempty"`

	// Hash contains the actual hash code of the signed document.
	Hash Sha3H256 `json:"sha3H256,omitempty"`
}

type SignData struct {
	// A human readable hint about the location where this has been signed. A [core.NavigationPath] may make sense.
	Location string

	// Internal Resources to resolve and hash. This will be resolved as a NamedStore.
	Resources []user.Resource

	// Optional image reference to a handwritten image of the signature. This has no legal worth.
	SignatureImage image.ID
}

// SignUnqualifiedWithSubject treats the given subject as the signer authority. This will only succeed if
// the subject is valid. The result is a simple and unqualified signature. It will usually not be accepted by
// law as a full qualified electronic signature. However, user subjects must be authenticated at least by
// confirming its mail address. There is no extra permission required to sign something.
type SignUnqualifiedWithSubject func(subject user.Subject, data SignData) (ID, error)

type AnonSignData struct {
	Firstname string
	Lastname  string
	Email     user.Email
	SignData
}

// SignUnqualified creates a new signature without a subject. This must not be used if the subject is known
// and authenticated. This is for workflows, where the workflow does not require an authenticated user. This
// is common where a device is given to a human who must just sign or confirm a report or for unqualified consents
// by mail.
type SignUnqualified func(signData AnonSignData) (ID, error)

// FindSignaturesByUser returns all Signatures that the given user has ever made. A signature cannot be removed.
// A subject can always get its own signatures.
type FindSignaturesByUser func(subject user.Subject, uid user.ID) iter.Seq2[Signature, error]

// FindSignaturesByResource returns all Signatures that the given resource has ever got. A subject can always
// get its own signatures.
type FindSignaturesByResource func(subject user.Subject, res user.Resource) iter.Seq2[Signature, error]

type FindByID func(subject user.Subject, id ID) (option.Opt[Signature], error)

// ValidateSignatureChain recalculates the signature hash based on the rules as defined by [Signature]. This
// does not validate the existence or the original bytes involved in calculating the hashes. A subject
// can always validate its own signatures.
type ValidateSignatureChain func(subject user.Subject, id ID) error

type UserSettings struct {
	User           user.ID
	ImageSignature image.ID
}

func (s UserSettings) Identity() user.ID {
	return s.User
}

type LoadUserSettings func(subject user.Subject, uid user.ID) (UserSettings, error)

type UpdateUserSettingsData struct {
	ImageSignature image.ID
}
type UpdateUserSettings func(subject user.Subject, uid user.ID, cdata UpdateUserSettingsData) error

type UserSettingsRepository data.Repository[UserSettings, user.ID]

type Repository data.Repository[Signature, ID]

type UseCases struct {
	SignUnqualifiedWithSubject SignUnqualifiedWithSubject
	SignUnqualified            SignUnqualified
	FindSignaturesByUser       FindSignaturesByUser
	FindSignaturesByResource   FindSignaturesByResource
	ValidateSignatureChain     ValidateSignatureChain
	FindByID                   FindByID
	LoadUserSettings           LoadUserSettings
	UpdateUserSettings         UpdateUserSettings
}

// NewUseCases loads and keeps the entire repository in memory and builds an inverse index so that the Find* use cases
// have efficient O(1) implementations.
func NewUseCases(stores blob.Stores, repo Repository, settingsRepo UserSettingsRepository, openImgReader image.OpenReader) (UseCases, error) {
	var mutex sync.Mutex
	var index inMemoryIndex

	for sig, err := range repo.All() {
		if err != nil {
			return UseCases{}, err
		}

		index.Index(sig)
	}

	return UseCases{
		SignUnqualified:            NewSignUnqualified(&mutex, openImgReader, repo, stores, &index),
		SignUnqualifiedWithSubject: NewSignUnqualifiedWithSubject(&mutex, openImgReader, repo, stores, &index),
		FindSignaturesByUser:       NewFindSignaturesByUser(&index),
		FindSignaturesByResource:   NewFindSignaturesByResource(&index),
		FindByID:                   NewFindByID(repo, &index),
		LoadUserSettings:           NewLoadUserSettings(settingsRepo),
		UpdateUserSettings:         NewUpdateUserSettings(settingsRepo),
	}, nil
}
