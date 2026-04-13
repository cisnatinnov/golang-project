package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/lib/pq"
)

// Note: These tests are designed to run with a real database
// Set DATABASE_URL environment variable to run tests
// Example: DATABASE_URL=postgres://postgres:postgres@localhost:5432/test_db go test ./repository

func getTestRepo(t *testing.T) *Repository {
	dbDsn := "postgres://postgres:postgres@db:5432/database?sslmode=disable"
	repo := NewRepository(NewRepositoryOptions{Dsn: dbDsn})
	if repo == nil || repo.Db == nil {
		t.Skip("Database connection failed - skipping integration tests")
	}
	return repo
}

// Test Person Email CRUD
func TestPersonEmailCRUD(t *testing.T) {
	repo := getTestRepo(t)
	ctx := context.Background()

	// Create a test user and person first
	userId := "test-user-" + "email"
	_, _ = repo.CreateUser(ctx, CreateUserInput{
		Username:     userId,
		PasswordHash: "test_hash",
	})

	_, _ = repo.CreatePerson(ctx, CreatePersonInput{
		UserId: userId,
	})

	t.Run("CreatePersonEmail", func(t *testing.T) {
		output, err := repo.CreatePersonEmail(ctx, CreatePersonEmailInput{
			UserId:    userId,
			Email:     "test@example.com",
			IsPrimary: true,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, output.Id)
	})

	t.Run("GetPersonEmailsByUserId", func(t *testing.T) {
		output, err := repo.GetPersonEmailsByUserId(ctx, GetPersonEmailsByUserIdInput{
			UserId: userId,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, output.Emails)
		assert.Equal(t, "test@example.com", output.Emails[0].Email)
	})

	t.Run("UpdatePersonEmail", func(t *testing.T) {
		emails, _ := repo.GetPersonEmailsByUserId(ctx, GetPersonEmailsByUserIdInput{UserId: userId})
		if len(emails.Emails) > 0 {
			verified := true
			err := repo.UpdatePersonEmail(ctx, UpdatePersonEmailInput{
				Id:       emails.Emails[0].Id,
				Verified: &verified,
			})
			assert.NoError(t, err)
		}
	})

	t.Run("DeletePersonEmail", func(t *testing.T) {
		emails, _ := repo.GetPersonEmailsByUserId(ctx, GetPersonEmailsByUserIdInput{UserId: userId})
		if len(emails.Emails) > 0 {
			err := repo.DeletePersonEmail(ctx, emails.Emails[0].Id)
			assert.NoError(t, err)
		}
	})

	// Cleanup
	_ = repo.DeletePerson(ctx, userId)
	_ = repo.DeleteUser(ctx, userId)
}

// Test Person Phone CRUD
func TestPersonPhoneCRUD(t *testing.T) {
	repo := getTestRepo(t)
	ctx := context.Background()

	// Create a test user and person first
	userId := "test-user-" + "phone"
	_, _ = repo.CreateUser(ctx, CreateUserInput{
		Username:     userId,
		PasswordHash: "test_hash",
	})

	_, _ = repo.CreatePerson(ctx, CreatePersonInput{
		UserId: userId,
	})

	t.Run("CreatePersonPhone", func(t *testing.T) {
		phoneType := "mobile"
		output, err := repo.CreatePersonPhone(ctx, CreatePersonPhoneInput{
			UserId:    userId,
			Phone:     "+1234567890",
			Type:      &phoneType,
			IsPrimary: true,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, output.Id)
	})

	t.Run("GetPersonPhonesByUserId", func(t *testing.T) {
		output, err := repo.GetPersonPhonesByUserId(ctx, GetPersonPhonesByUserIdInput{
			UserId: userId,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, output.Phones)
		assert.Equal(t, "+1234567890", output.Phones[0].Phone)
	})

	t.Run("UpdatePersonPhone", func(t *testing.T) {
		phones, _ := repo.GetPersonPhonesByUserId(ctx, GetPersonPhonesByUserIdInput{UserId: userId})
		if len(phones.Phones) > 0 {
			verified := true
			err := repo.UpdatePersonPhone(ctx, UpdatePersonPhoneInput{
				Id:       phones.Phones[0].Id,
				Verified: &verified,
			})
			assert.NoError(t, err)
		}
	})

	t.Run("DeletePersonPhone", func(t *testing.T) {
		phones, _ := repo.GetPersonPhonesByUserId(ctx, GetPersonPhonesByUserIdInput{UserId: userId})
		if len(phones.Phones) > 0 {
			err := repo.DeletePersonPhone(ctx, phones.Phones[0].Id)
			assert.NoError(t, err)
		}
	})

	// Cleanup
	_ = repo.DeletePerson(ctx, userId)
	_ = repo.DeleteUser(ctx, userId)
}

// Test Person Social Media CRUD
func TestPersonSocialMediaCRUD(t *testing.T) {
	repo := getTestRepo(t)
	ctx := context.Background()

	// Create a test user and person first
	userId := "test-user-" + "social"
	_, _ = repo.CreateUser(ctx, CreateUserInput{
		Username:     userId,
		PasswordHash: "test_hash",
	})

	_, _ = repo.CreatePerson(ctx, CreatePersonInput{
		UserId: userId,
	})

	t.Run("CreatePersonSocialMedia", func(t *testing.T) {
		username := "testuser"
		output, err := repo.CreatePersonSocialMedia(ctx, CreatePersonSocialMediaInput{
			UserId:   userId,
			Platform: "twitter",
			Username: &username,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, output.Id)
	})

	t.Run("GetPersonSocialMediaByUserId", func(t *testing.T) {
		output, err := repo.GetPersonSocialMediaByUserId(ctx, GetPersonSocialMediaByUserIdInput{
			UserId: userId,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, output.SocialMediaAccounts)
		assert.Equal(t, "twitter", output.SocialMediaAccounts[0].Platform)
	})

	t.Run("UpdatePersonSocialMedia", func(t *testing.T) {
		socials, _ := repo.GetPersonSocialMediaByUserId(ctx, GetPersonSocialMediaByUserIdInput{UserId: userId})
		if len(socials.SocialMediaAccounts) > 0 {
			newUrl := "https://twitter.com/newuser"
			err := repo.UpdatePersonSocialMedia(ctx, UpdatePersonSocialMediaInput{
				Id:         socials.SocialMediaAccounts[0].Id,
				ProfileUrl: &newUrl,
			})
			assert.NoError(t, err)
		}
	})

	t.Run("DeletePersonSocialMedia", func(t *testing.T) {
		socials, _ := repo.GetPersonSocialMediaByUserId(ctx, GetPersonSocialMediaByUserIdInput{UserId: userId})
		if len(socials.SocialMediaAccounts) > 0 {
			err := repo.DeletePersonSocialMedia(ctx, socials.SocialMediaAccounts[0].Id)
			assert.NoError(t, err)
		}
	})

	// Cleanup
	_ = repo.DeletePerson(ctx, userId)
	_ = repo.DeleteUser(ctx, userId)
}

// Test Full User Profile Creation Flow
func TestFullUserProfileFlow(t *testing.T) {
	repo := getTestRepo(t)
	ctx := context.Background()

	userId := "full-flow-user"

	t.Run("Complete User Profile Creation", func(t *testing.T) {
		// 1. Create user
		userOut, err := repo.CreateUser(ctx, CreateUserInput{
			Username:     userId,
			PasswordHash: "hashed_password",
		})
		require.NoError(t, err)
		require.NotEmpty(t, userOut.Id)

		// 2. Create person profile
		personOut, err := repo.CreatePerson(ctx, CreatePersonInput{
			UserId:    userOut.Id,
			FirstName: stringPtr("John"),
			LastName:  stringPtr("Doe"),
		})
		require.NoError(t, err)
		require.NotEmpty(t, personOut.Id)

		// 3. Add emails
		emailOut, err := repo.CreatePersonEmail(ctx, CreatePersonEmailInput{
			UserId:    userOut.Id,
			Email:     "john@example.com",
			IsPrimary: true,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, emailOut.Id)

		// 4. Add phones
		phoneType := "mobile"
		phoneOut, err := repo.CreatePersonPhone(ctx, CreatePersonPhoneInput{
			UserId:    userOut.Id,
			Phone:     "+11234567890",
			Type:      &phoneType,
			IsPrimary: true,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, phoneOut.Id)

		// 5. Add social media
		username := "johndoe"
		socialOut, err := repo.CreatePersonSocialMedia(ctx, CreatePersonSocialMediaInput{
			UserId:   userOut.Id,
			Platform: "github",
			Username: &username,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, socialOut.Id)

		// 6. Verify all data is retrievable
		person, err := repo.GetPersonByUserId(ctx, GetPersonByUserIdInput{UserId: userOut.Id})
		assert.NoError(t, err)
		assert.Equal(t, *person.FirstName, "John")
		assert.Equal(t, *person.LastName, "Doe")

		emails, err := repo.GetPersonEmailsByUserId(ctx, GetPersonEmailsByUserIdInput{UserId: userOut.Id})
		assert.NoError(t, err)
		assert.Len(t, emails.Emails, 1)
		assert.Equal(t, emails.Emails[0].Email, "john@example.com")

		phones, err := repo.GetPersonPhonesByUserId(ctx, GetPersonPhonesByUserIdInput{UserId: userOut.Id})
		assert.NoError(t, err)
		assert.Len(t, phones.Phones, 1)
		assert.Equal(t, phones.Phones[0].Phone, "+11234567890")

		socials, err := repo.GetPersonSocialMediaByUserId(ctx, GetPersonSocialMediaByUserIdInput{UserId: userOut.Id})
		assert.NoError(t, err)
		assert.Len(t, socials.SocialMediaAccounts, 1)
		assert.Equal(t, socials.SocialMediaAccounts[0].Platform, "github")

		// Cleanup
		_ = repo.DeletePersonSocialMedia(ctx, socialOut.Id)
		_ = repo.DeletePersonPhone(ctx, phoneOut.Id)
		_ = repo.DeletePersonEmail(ctx, emailOut.Id)
		_ = repo.DeletePerson(ctx, personOut.Id)
		_ = repo.DeleteUser(ctx, userOut.Id)
	})
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
