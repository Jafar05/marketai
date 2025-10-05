package postgres

const (
	getByUserName = `
		SELECT 
			id, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE email=$1
	`

	createUser = `
		INSERT INTO users
			(id, email, password_hash, phoneNumber, role, created_at, updated_at)
        VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)
		RETURNING id`
)
