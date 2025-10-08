package postgres

const (
	getByUserName = `
		SELECT 
			id, email, password_hash, role, email_verified, created_at, updated_at
		FROM users
		WHERE email=$1 AND phone_number=$2
	`

	createUser = `
		INSERT INTO users
			(id, full_name, email, password_hash, phone_number, role, created_at, updated_at)
        VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	createToken = `
		INSERT INTO email_verification_tokens (user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
	`

	getUserByToken = `
		SELECT user_id FROM email_verification_tokens 
		WHERE token=$1 AND expires_at > NOW()
	`

	deleteToken = `DELETE FROM email_verification_tokens WHERE token=$1`
)
