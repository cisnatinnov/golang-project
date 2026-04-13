package repository

import (
	"context"

	"github.com/google/uuid"
)

func (r *Repository) CreateEstate(ctx context.Context, input CreateEstateInput) (output CreateEstateOutput, err error) {
	id := uuid.New().String()
	_, err = r.Db.ExecContext(ctx, "INSERT INTO estate (id, length, width) VALUES ($1, $2, $3)", id, input.Length, input.Width)
	if err != nil {
		return
	}
	output.Id = id
	return
}

func (r *Repository) GetEstateById(ctx context.Context, id string) (estate Estate, err error) {
	err = r.Db.QueryRowContext(ctx, "SELECT id, length, width FROM estate WHERE id = $1", id).Scan(&estate.Id, &estate.Length, &estate.Width)
	return
}

func (r *Repository) CreateTree(ctx context.Context, input CreateTreeInput) (output CreateTreeOutput, err error) {
	id := uuid.New().String()
	_, err = r.Db.ExecContext(ctx, "INSERT INTO tree (id, estate_id, x, y, height) VALUES ($1, $2, $3, $4, $5)", id, input.EstateId, input.X, input.Y, input.Height)
	if err != nil {
		return
	}
	output.Id = id
	return
}

func (r *Repository) GetEstateStats(ctx context.Context, input GetEstateStatsInput) (output GetEstateStatsOutput, err error) {
	err = r.Db.QueryRowContext(ctx, `
		SELECT
			COUNT(id),
			COALESCE(MAX(height), 0),
			COALESCE(MIN(height), 0),
			COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY height), 0)
		FROM tree
		WHERE estate_id = $1
	`, input.EstateId).Scan(&output.Count, &output.Max, &output.Min, &output.Median)
	return
}

func (r *Repository) GetTreesByEstateId(ctx context.Context, input GetTreesByEstateIdInput) (output GetTreesByEstateIdOutput, err error) {
	rows, err := r.Db.QueryContext(ctx, "SELECT id, estate_id, x, y, height FROM tree WHERE estate_id = $1", input.EstateId)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var t Tree
		if err = rows.Scan(&t.Id, &t.EstateId, &t.X, &t.Y, &t.Height); err != nil {
			return
		}
		output.Trees = append(output.Trees, t)
	}
	err = rows.Err()
	return
}

func (r *Repository) CreateUser(ctx context.Context, input CreateUserInput) (output CreateUserOutput, err error) {
	id := uuid.New().String()
	_, err = r.Db.ExecContext(ctx, "INSERT INTO users (id, username, password_hash) VALUES ($1, $2, $3)", id, input.Username, input.PasswordHash)
	if err != nil {
		return
	}
	output.Id = id
	return
}

func (r *Repository) GetUserById(ctx context.Context, id string) (user User, err error) {
	err = r.Db.QueryRowContext(ctx, "SELECT id, username, password_hash FROM users WHERE id = $1", id).Scan(&user.Id, &user.Username, &user.PasswordHash)
	return
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (user User, err error) {
	err = r.Db.QueryRowContext(ctx, "SELECT id, username, password_hash FROM users WHERE username = $1", username).Scan(&user.Id, &user.Username, &user.PasswordHash)
	return
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (user User, err error) {
	err = r.Db.QueryRowContext(ctx, "SELECT u.id, u.username, u.password_hash FROM users u JOIN person_email pe ON u.id = pe.user_id WHERE pe.email = $1", email).Scan(&user.Id, &user.Username, &user.PasswordHash)
	return
}

func (r *Repository) GetUserByUsernameOrEmail(ctx context.Context, input GetUserByUsernameOrEmailInput) (user User, err error) {
	err = r.Db.QueryRowContext(ctx, `
		SELECT u.id, u.username, u.password_hash FROM users u
		LEFT JOIN person_email pe ON u.id = pe.user_id
		WHERE (u.username = $1 AND $1 != '') OR (pe.email = $2 AND $2 != '')
		LIMIT 1
	`, input.Username, input.Email).Scan(&user.Id, &user.Username, &user.PasswordHash)
	return
}

func (r *Repository) UpdateUser(ctx context.Context, input UpdateUserInput) (err error) {
	_, err = r.Db.ExecContext(ctx, "UPDATE users SET username = $1, password_hash = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3", input.Username, input.PasswordHash, input.Id)
	return
}

func (r *Repository) DeleteUser(ctx context.Context, id string) (err error) {
	_, err = r.Db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	return
}

func (r *Repository) CreatePerson(ctx context.Context, input CreatePersonInput) (output CreatePersonOutput, err error) {
	id := uuid.New().String()
	_, err = r.Db.ExecContext(ctx, `
		INSERT INTO person (id, user_id, first_name, last_name, date_of_birth, bio, avatar_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, input.UserId, input.FirstName, input.LastName, input.DateOfBirth, input.Bio, input.AvatarUrl)
	if err != nil {
		return
	}
	output.Id = id
	return
}

func (r *Repository) GetPersonByUserId(ctx context.Context, input GetPersonByUserIdInput) (person Person, err error) {
	err = r.Db.QueryRowContext(ctx, `
		SELECT id, user_id, first_name, last_name, date_of_birth, bio, avatar_url
		FROM person
		WHERE user_id = $1
	`, input.UserId).Scan(&person.Id, &person.UserId, &person.FirstName, &person.LastName, &person.DateOfBirth, &person.Bio, &person.AvatarUrl)
	return
}

func (r *Repository) UpdatePerson(ctx context.Context, input UpdatePersonInput) (err error) {
	_, err = r.Db.ExecContext(ctx, `
		UPDATE person
		SET first_name = COALESCE($1, first_name),
		    last_name = COALESCE($2, last_name),
		    date_of_birth = COALESCE($3, date_of_birth),
		    bio = COALESCE($4, bio),
		    avatar_url = COALESCE($5, avatar_url),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
	`, input.FirstName, input.LastName, input.DateOfBirth, input.Bio, input.AvatarUrl, input.Id)
	return
}

func (r *Repository) DeletePerson(ctx context.Context, id string) (err error) {
	_, err = r.Db.ExecContext(ctx, "DELETE FROM person WHERE id = $1", id)
	return
}

// Email CRUD
func (r *Repository) CreatePersonEmail(ctx context.Context, input CreatePersonEmailInput) (output CreatePersonEmailOutput, err error) {
	id := uuid.New().String()
	_, err = r.Db.ExecContext(ctx, `
		INSERT INTO person_email (id, user_id, email, is_primary)
		VALUES ($1, $2, $3, $4)
	`, id, input.UserId, input.Email, input.IsPrimary)
	if err != nil {
		return
	}
	output.Id = id
	return
}

func (r *Repository) GetPersonEmailsByUserId(ctx context.Context, input GetPersonEmailsByUserIdInput) (output GetPersonEmailsByUserIdOutput, err error) {
	rows, err := r.Db.QueryContext(ctx, `
		SELECT id, user_id, email, is_primary, verified
		FROM person_email
		WHERE user_id = $1
		ORDER BY is_primary DESC, created_at ASC
	`, input.UserId)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var email PersonEmail
		if err = rows.Scan(&email.Id, &email.UserId, &email.Email, &email.IsPrimary, &email.Verified); err != nil {
			return
		}
		output.Emails = append(output.Emails, email)
	}
	err = rows.Err()
	return
}

func (r *Repository) UpdatePersonEmail(ctx context.Context, input UpdatePersonEmailInput) (err error) {
	_, err = r.Db.ExecContext(ctx, `
		UPDATE person_email
		SET is_primary = COALESCE($1, is_primary),
		    verified = COALESCE($2, verified),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`, input.IsPrimary, input.Verified, input.Id)
	return
}

func (r *Repository) DeletePersonEmail(ctx context.Context, id string) (err error) {
	_, err = r.Db.ExecContext(ctx, "DELETE FROM person_email WHERE id = $1", id)
	return
}

// Phone CRUD
func (r *Repository) CreatePersonPhone(ctx context.Context, input CreatePersonPhoneInput) (output CreatePersonPhoneOutput, err error) {
	id := uuid.New().String()
	_, err = r.Db.ExecContext(ctx, `
		INSERT INTO person_phone (id, user_id, phone, type, is_primary)
		VALUES ($1, $2, $3, $4, $5)
	`, id, input.UserId, input.Phone, input.Type, input.IsPrimary)
	if err != nil {
		return
	}
	output.Id = id
	return
}

func (r *Repository) GetPersonPhonesByUserId(ctx context.Context, input GetPersonPhonesByUserIdInput) (output GetPersonPhonesByUserIdOutput, err error) {
	rows, err := r.Db.QueryContext(ctx, `
		SELECT id, user_id, phone, type, is_primary, verified
		FROM person_phone
		WHERE user_id = $1
		ORDER BY is_primary DESC, created_at ASC
	`, input.UserId)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var phone PersonPhone
		if err = rows.Scan(&phone.Id, &phone.UserId, &phone.Phone, &phone.Type, &phone.IsPrimary, &phone.Verified); err != nil {
			return
		}
		output.Phones = append(output.Phones, phone)
	}
	err = rows.Err()
	return
}

func (r *Repository) UpdatePersonPhone(ctx context.Context, input UpdatePersonPhoneInput) (err error) {
	_, err = r.Db.ExecContext(ctx, `
		UPDATE person_phone
		SET type = COALESCE($1, type),
		    is_primary = COALESCE($2, is_primary),
		    verified = COALESCE($3, verified),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`, input.Type, input.IsPrimary, input.Verified, input.Id)
	return
}

func (r *Repository) DeletePersonPhone(ctx context.Context, id string) (err error) {
	_, err = r.Db.ExecContext(ctx, "DELETE FROM person_phone WHERE id = $1", id)
	return
}

// Social Media CRUD
func (r *Repository) CreatePersonSocialMedia(ctx context.Context, input CreatePersonSocialMediaInput) (output CreatePersonSocialMediaOutput, err error) {
	id := uuid.New().String()
	_, err = r.Db.ExecContext(ctx, `
		INSERT INTO person_social_media (id, user_id, platform, username, profile_url)
		VALUES ($1, $2, $3, $4, $5)
	`, id, input.UserId, input.Platform, input.Username, input.ProfileUrl)
	if err != nil {
		return
	}
	output.Id = id
	return
}

func (r *Repository) GetPersonSocialMediaByUserId(ctx context.Context, input GetPersonSocialMediaByUserIdInput) (output GetPersonSocialMediaByUserIdOutput, err error) {
	rows, err := r.Db.QueryContext(ctx, `
		SELECT id, user_id, platform, username, profile_url
		FROM person_social_media
		WHERE user_id = $1
		ORDER BY created_at ASC
	`, input.UserId)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var sm PersonSocialMedia
		if err = rows.Scan(&sm.Id, &sm.UserId, &sm.Platform, &sm.Username, &sm.ProfileUrl); err != nil {
			return
		}
		output.SocialMediaAccounts = append(output.SocialMediaAccounts, sm)
	}
	err = rows.Err()
	return
}

func (r *Repository) UpdatePersonSocialMedia(ctx context.Context, input UpdatePersonSocialMediaInput) (err error) {
	_, err = r.Db.ExecContext(ctx, `
		UPDATE person_social_media
		SET username = COALESCE($1, username),
		    profile_url = COALESCE($2, profile_url),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`, input.Username, input.ProfileUrl, input.Id)
	return
}

func (r *Repository) DeletePersonSocialMedia(ctx context.Context, id string) (err error) {
	_, err = r.Db.ExecContext(ctx, "DELETE FROM person_social_media WHERE id = $1", id)
	return
}
