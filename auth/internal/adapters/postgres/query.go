package postgres

const (
	getByUserName = `
		SELECT 
			id, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE email=$1 AND phone_number=$2
	`

	createUser = `
		INSERT INTO users
			(id, full_name, email, password_hash, phone_number, role, created_at, updated_at)
        VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
)
