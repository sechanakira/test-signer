package db

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestInitDB(t *testing.T) {
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_PORT", "testport")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpassword")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SSL_MODE", "disable")

	expectedDSN := "host=testhost port=testport user=testuser password=testpassword dbname=testdb sslmode=disable"
	dsn := getDSN()

	assert.Equal(t, expectedDSN, dsn, "DSN should match the expected format")
}
