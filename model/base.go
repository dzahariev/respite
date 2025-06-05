package model

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Object is an abstration of all Base objects
type Object interface {
	ResourceName() string
	IsGlobal() bool
	GetID() uuid.UUID
	SetID(uuid.UUID)
	GetCreatedAt() *time.Time
	GetUpdatedAt() *time.Time
	Save(ctx context.Context, db *gorm.DB, object Object) error
	Count(ctx context.Context, db *gorm.DB, object Object) (int64, error)
	FindAll(ctx context.Context, db *gorm.DB, object Object) (*[]Object, error)
	FindByID(ctx context.Context, db *gorm.DB, object Object, uid uuid.UUID) error
	Update(ctx context.Context, db *gorm.DB, object Object) error
	Delete(ctx context.Context, db *gorm.DB, object Object) error
	Prepare(ctx context.Context) error
	Validate(ctx context.Context) error
	Preloads() []string
}

// LocalObject is an abstration of an Local objects
type LocalObject interface {
	SetUserID(uuid.UUID)
}

// Base holds technical fields
type Base struct {
	ID        uuid.UUID  `json:"id"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// IsGlobal returns the global flag
func (b *Base) IsGlobal() bool {
	return false
}

// GetID returns the ID
func (b *Base) GetID() uuid.UUID {
	return b.ID
}

// SetID returns the ID
func (b *Base) SetID(id uuid.UUID) {
	b.ID = id
}

// GetCreatedAt returns the CreatedAt
func (b *Base) GetCreatedAt() *time.Time {
	return b.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt
func (b *Base) GetUpdatedAt() *time.Time {
	return b.UpdatedAt
}

// Preloads returns the preloads
func (b *Base) Preloads() []string {
	return []string{}
}

// Prepare is a hook that is called before saving
func (b *Base) Prepare() error {
	return nil
}

// Validate checks structure consistency
func (b *Base) Validate() error {
	return nil
}

// BasePrepare initilises techncal fields
func (b *Base) BasePrepare() error {
	if b.ID.IsNil() {
		uuid, err := uuid.NewV4()
		if err != nil {
			return err
		}

		b.ID = uuid
	}
	return nil
}

// Count returns count of all known objects of this type
func (b *Base) Count(ctx context.Context, db *gorm.DB, object Object) (int64, error) {
	var count int64
	err := db.Model(object).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// FindByID returns an objects with corresponding ID if exists
func (b *Base) FindByID(ctx context.Context, db *gorm.DB, object Object, uid uuid.UUID) error {
	preloads := object.Preloads()
	if len(preloads) != 0 {
		for _, preload := range preloads {
			db.Preload(preload)
		}
	} else {
		db.Preload(clause.Associations)
	}
	err := db.Model(object).First(object, uid).Error
	if err != nil {
		return err
	}
	return nil
}

// FindAll returns all known objects of this type
func (b *Base) FindAll(ctx context.Context, db *gorm.DB, object Object) (*[]Object, error) {
	entites := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(object)), 0, 0).Interface()
	preloads := object.Preloads()
	if len(preloads) != 0 {
		for _, preload := range preloads {
			db.Preload(preload)
		}
	} else {
		db.Preload(clause.Associations)
	}
	err := db.Model(&object).Find(&entites).Error
	if err != nil {
		return &[]Object{}, err
	}

	objects := []Object{}
	s := reflect.ValueOf(entites)
	for i := 0; i < s.Len(); i++ {
		currentEntity := s.Index(i).Interface().(Object)
		objects = append(objects, currentEntity)
	}
	return &objects, nil
}

// Delete is removing existing objects
func (b *Base) Delete(ctx context.Context, db *gorm.DB, object Object) error {
	err := db.Delete(object).Error
	if err != nil {
		return err
	}
	return nil
}

// Save saves the structure as new object
func (b *Base) Save(ctx context.Context, db *gorm.DB, object Object) error {
	err := object.Prepare(ctx)
	if err != nil {
		return err
	}

	err = object.Validate(ctx)
	if err != nil {
		return err
	}

	err = db.Create(object).Error
	if err != nil {
		return err
	}
	return nil
}

// Update updates the existing objects
func (b *Base) Update(ctx context.Context, db *gorm.DB, object Object) error {
	if b.ID == uuid.Nil {
		return fmt.Errorf("cannot update non saved entity")
	}

	err := object.Validate(ctx)
	if err != nil {
		return err
	}

	err = db.Updates(object).Error

	if err != nil {
		return err
	}
	return nil
}
