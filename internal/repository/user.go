package repository

import (
	"errors"
	"go-auth/pkg/logger"
	"time"
)

// User: Kullanıcı modeli
type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // JSON'da gösterilmez
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRepository: User repository arayüzü
type UserRepository interface {
	Create(user *User) error
	GetByID(id int64) (*User, error)
	GetByEmail(email string) (*User, error)
	Update(user *User) error
	Delete(id int64) error
}

// userRepository: User repository implementasyonu
type userRepository struct {
	logger logger.Logger
	// Gelecekte database bağlantısı buraya eklenecek
	// db *sql.DB veya *gorm.DB
	users map[int64]*User // Basit in-memory storage (örnek amaçlı)
}

// NewUserRepository: User repository constructor'ı
// Logger dependency injection ile enjekte edilir
func NewUserRepository(log logger.Logger) UserRepository {
	repo := &userRepository{
		logger: log,
		users:  make(map[int64]*User),
	}

	repo.logger.Info("User repository oluşturuldu")
	return repo
}

// Create: Yeni kullanıcı oluşturur
func (r *userRepository) Create(user *User) error {
	r.logger.Debug("Kullanıcı oluşturuluyor", "email", user.Email)

	// Email kontrolü
	if _, exists := r.getByEmail(user.Email); exists {
		r.logger.Warn("Email zaten kullanılıyor", "email", user.Email)
		return errors.New("email already exists")
	}

	// ID atama (gerçek uygulamada database auto-increment kullanılır)
	user.ID = int64(len(r.users) + 1)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	r.users[user.ID] = user

	r.logger.Info("Kullanıcı başarıyla oluşturuldu", "id", user.ID, "email", user.Email)
	return nil
}

// GetByID: ID'ye göre kullanıcı getirir
func (r *userRepository) GetByID(id int64) (*User, error) {
	r.logger.Debug("Kullanıcı aranıyor", "id", id)

	user, exists := r.users[id]
	if !exists {
		r.logger.Warn("Kullanıcı bulunamadı", "id", id)
		return nil, errors.New("user not found")
	}

	r.logger.Debug("Kullanıcı bulundu", "id", id, "email", user.Email)
	return user, nil
}

// GetByEmail: Email'e göre kullanıcı getirir
func (r *userRepository) GetByEmail(email string) (*User, error) {
	r.logger.Debug("Email ile kullanıcı aranıyor", "email", email)

	user, exists := r.getByEmail(email)
	if !exists {
		r.logger.Warn("Email ile kullanıcı bulunamadı", "email", email)
		return nil, errors.New("user not found")
	}

	r.logger.Debug("Email ile kullanıcı bulundu", "email", email, "id", user.ID)
	return user, nil
}

// getByEmail: Internal helper method
func (r *userRepository) getByEmail(email string) (*User, bool) {
	for _, user := range r.users {
		if user.Email == email {
			return user, true
		}
	}
	return nil, false
}

// Update: Kullanıcı bilgilerini günceller
func (r *userRepository) Update(user *User) error {
	r.logger.Debug("Kullanıcı güncelleniyor", "id", user.ID)

	existingUser, exists := r.users[user.ID]
	if !exists {
		r.logger.Warn("Güncellenecek kullanıcı bulunamadı", "id", user.ID)
		return errors.New("user not found")
	}

	// Email değişikliği kontrolü
	if user.Email != existingUser.Email {
		if _, emailExists := r.getByEmail(user.Email); emailExists {
			r.logger.Warn("Yeni email zaten kullanılıyor", "email", user.Email)
			return errors.New("email already exists")
		}
	}

	user.UpdatedAt = time.Now()
	r.users[user.ID] = user

	r.logger.Info("Kullanıcı başarıyla güncellendi", "id", user.ID)
	return nil
}

// Delete: Kullanıcıyı siler
func (r *userRepository) Delete(id int64) error {
	r.logger.Debug("Kullanıcı siliniyor", "id", id)

	_, exists := r.users[id]
	if !exists {
		r.logger.Warn("Silinecek kullanıcı bulunamadı", "id", id)
		return errors.New("user not found")
	}

	delete(r.users, id)

	r.logger.Info("Kullanıcı başarıyla silindi", "id", id)
	return nil
}
